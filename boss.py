import asyncio
import json
import math
import os
import platform
from string import Template
from typing import Dict, Optional

import asyncio_redis
import prometheus_client
from aiohttp import web
from jinja2 import Environment, FileSystemLoader, select_autoescape

import util
import ws
from config import THE_HEADS_EVENTS, get_args, Config
from consul_config import ConsulBackend
from etcd_config import lock
from grid import the_grid
from head_manager import HeadManager
from health import health_check
from installation import build_installation, Installation
from metrics import handle_metrics
from orchestrator import Orchestrator
from transformations import Mat, Vec

REDIS_MESSAGE_RECEIVED = prometheus_client.Counter(
    "heads_boss_redis_message_ingested",
    "Message ingested from redis",
    ["channel", "type", "src"],
)

TASKS = prometheus_client.Gauge(
    "heads_boss_tasks",
    "Number of asyncio tasks",
    [],
)

TASKS.set_function(lambda: len(asyncio.Task.all_tasks()))


async def home(request):
    cfg = request.app['cfg']['cfg']

    keys = await cfg.get_keys("/the-heads/installation/")

    installations = set(k.split("/")[2] for k in keys)

    jinja_env = request.app['jinja_env']

    template = jinja_env.get_template('boss.html')

    hostname = platform.node()
    result = template.render(installations=installations, hostname=hostname)

    return web.Response(text=result, content_type="text/html")


async def fetch(session, url):
    async with session.get(url) as response:
        return await response.text()


async def head_positioned(
        inst: Installation,
        ws_manager: ws.WebsocketManager,
        head_manager: HeadManager,
        msg: Dict
):
    print(msg)
    ws_manager.send(msg)


async def motion_detected(
        inst: Installation,
        ws_manager: ws.WebsocketManager,
        head_manager: HeadManager,
        orchestrator: Orchestrator,
        msg: Dict
):
    data = msg['data']

    cam = inst.cameras[data['cameraName']]
    p0 = Vec(0, 0)
    p1 = Mat.rotz(data['position']) * Vec(5, 0)

    p0 = cam.stand.m * cam.m * p0
    p1 = cam.stand.m * cam.m * p1

    step_size = min(the_grid.get_pixel_size()) / 4.0

    to = p1 - p0
    length = (to).abs()
    direction = to.scale(1.0 / length)

    dx = to.x / length * step_size
    dy = to.y / length * step_size

    initial = p0 + direction.scale(0.5)
    pos_x, pos_y = initial.x, initial.y

    steps = int(length / step_size)
    for i in range(steps):
        prev_xy = the_grid.get(cam, pos_x, pos_y)
        if prev_xy is None:
            break
        the_grid.set(cam, pos_x, pos_y, prev_xy + 0.025)
        pos_x += dx
        pos_y += dy

    focus = Vec(*the_grid.idx_to_xy(the_grid.focus()))
    orchestrator.focus = focus
    orchestrator.act()


async def run_redis(redis_hostport, ws_manager, inst: Installation, hm: HeadManager, orchestrator: Orchestrator):
    print("Connecting to redis:", redis_hostport)
    host, port = redis_hostport.split(":")
    connection = await asyncio_redis.Connection.create(host=host, port=int(port))
    print("Connected to redis", redis_hostport)
    subscriber = await connection.start_subscribe()
    await subscriber.subscribe([THE_HEADS_EVENTS])

    while True:
        reply = await subscriber.next_published()
        msg = json.loads(reply.value)

        data = msg['data']
        src = data.get('cameraName') or data.get('headName') or data['name']

        REDIS_MESSAGE_RECEIVED.labels(
            reply.channel,
            msg['type'],
            src,
        ).inc()

        if msg['type'] == "motion-detected":
            await motion_detected(inst, ws_manager, hm, orchestrator, msg)

        if msg['type'] == "head-positioned":
            await head_positioned(inst, ws_manager, hm, msg)

        if msg['type'] == "active":
            ws_manager.send(msg)


async def html_handler(request):
    filename = request.match_info.get('name') + ".html"
    with open(filename) as fp:
        contents = Template(fp.read())

    text = contents.safe_substitute()
    return web.Response(text=text, content_type="text/html")


def static_text_handler(extension):
    # TODO: make sure .. not allowed in paths, etc.
    content_type = {
        "js": "text/javascript",
    }[extension]

    async def handler(request):
        filename = request.match_info.get('name') + "." + extension
        with open(filename) as fp:
            text = fp.read()
        return web.Response(text=text, content_type=content_type)

    return handler


def static_binary_handler(extension):
    # TODO: make sure .. not allowed in paths, etc.
    content_type = {
        "png": "image/png",
    }[extension]

    async def handler(request):
        filename = request.match_info.get('name') + "." + extension
        with open(filename, "rb") as fp:
            body = fp.read()
        return web.Response(body=body, content_type=content_type)

    return handler


def random_png(request):
    # width, height = 256 * 2, 256 * 2
    #
    # t0 = time.time()
    #
    # a = np.zeros((height, width, 4), dtype=np.uint8)
    # a[..., 1] = np.random.randint(50, 200, size=(height, width), dtype=np.uint8)
    # a[..., 3] = 255
    #
    # print(time.time() - t0)
    #
    # t0 = time.time()
    # body = png.write_png(a.tobytes(), width, height)
    # print(time.time() - t0)
    #
    # print(len(body))
    body = the_grid.to_png()

    return web.Response(body=body, content_type="image/png")


async def installation_handler(request):
    name = request.match_info.get('installation')
    cfg = request.app['cfg']

    result = await build_installation(name, cfg['cfg'])

    return web.Response(text=json.dumps(result), content_type="application/json")


async def task_handler(request):
    tasks = list(sorted(asyncio.Task.all_tasks(), key=id))
    text = ["tasks: {}".format(len(tasks))]
    for task in tasks:
        text.append("{} {}\n{}".format(
            hex(id(task)),
            task._state,
            str(task)
        ))

    return web.Response(text="\n\n".join(text), content_type="text/plain")


async def get_config(
        installation: str,
        port: int,
        config_endpoint: str,
):
    endpoint = ConsulBackend(config_endpoint)
    cfg = await Config(endpoint).setup(
        instance_name="boss-00",
        installation_override=installation,
    )

    resp, text = await endpoint.get_nodes_for_service("redis")
    assert resp.status == 200
    msg = json.loads(text)

    redis_servers = ["{}:{}".format(r['Address'], r['ServicePort']) for r in msg]

    assert len(redis_servers) > 0, "Need at least one redis server, for now"

    result = dict(
        endpoint=config_endpoint,
        installation=cfg.installation,
        redis_servers=redis_servers,
        cfg=cfg,
        port=port,
    )

    print("Using installation:", result['installation'])
    return result


async def aquire_lock(cfg):
    lockname = "/the-heads/installation/{installation}/boss/lock".format(**cfg)
    return await lock(cfg['endpoint'], lockname)


def frontend_handler(*path_prefix):
    async def handler(request):
        filename = request.match_info.get('filename')
        path = os.path.join(*path_prefix, filename)

        ext = os.path.splitext(path)[-1]

        mode = {".png": "rb"}.get(ext, "r")

        print(path)
        with open(path, mode) as fp:
            content = fp.read()

        content_type = {
            ".css": "text/css",
            ".json": "application/json",
            ".js": "text/javascript",
            ".map": "application/octet-stream",
            ".png": "image/png",
            ".html": "text/html",
        }[ext]

        if mode == "rb":
            return web.Response(body=content, content_type=content_type)
        else:
            return web.Response(text=content, content_type=content_type)

    return handler


async def setup(
        installation: str,
        port: int,
        config_endpoint: Optional[str] = "http://127.0.0.1:8500",
):
    cfg = await get_config(installation, port, config_endpoint)

    app = web.Application()
    app['cfg'] = cfg

    jinja_env = Environment(
        loader=FileSystemLoader('templates'),
        autoescape=select_autoescape(['html', 'xml'])
    )

    app['jinja_env'] = jinja_env

    ws_manager = ws.WebsocketManager()

    asyncio.ensure_future(the_grid.decay())

    json_inst = await build_installation(cfg['installation'], cfg['cfg'])
    inst = Installation.unmarshal(json_inst)

    app['inst'] = inst
    hm = HeadManager()

    app['head_manager'] = hm

    app.add_routes([
        web.get('/', home),
        web.get('/health', health_check),
        web.get('/metrics', handle_metrics),
        web.get('/ws', ws_manager.websocket_handler),
        web.get('/installation/{installation}/scene.json', installation_handler),
        web.get('/installation/{installation}/{name}.html', html_handler),
        web.get('/installation/{installation}/{name}.js', static_text_handler("js")),
        web.get('/installation/{installation}/{seed}/random.png', random_png),
        web.get('/installation/{installation}/{name}.png', static_binary_handler("png")),
        web.get("/tasks", task_handler),

        # Jenkins' frontend
        web.get("/build/{filename}", frontend_handler("boss-ui/build")),
        web.get("/build/json/{filename}", frontend_handler("boss-ui/build/json")),
        web.get("/build/media/{filename}", frontend_handler("boss-ui/build/media")),
        web.get("/build/js/{filename}", frontend_handler("boss-ui/build/js")),
        web.get("/static/js/{filename}", frontend_handler("boss-ui/build/static/js")),
        web.get("/static/css/{filename}", frontend_handler("boss-ui/build/static/css")),
    ])

    orchestrator = Orchestrator(
        inst=inst,
        ws_manager=ws_manager,
        head_manager=hm,
    )

    for redis in cfg['redis_servers']:
        asyncio.ensure_future(run_redis(redis, ws_manager, inst, hm, orchestrator))

    app['orchestrator'] = orchestrator

    return app


def main():
    args = get_args()

    loop = asyncio.get_event_loop()

    app = loop.run_until_complete(setup(
        installation=args.installation,
        config_endpoint=args.config_endpoint,
        port=args.port,
    ))

    loop.run_until_complete(util.run_app(app))
    loop.run_forever()


if __name__ == '__main__':
    main()
