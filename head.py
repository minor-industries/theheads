import argparse
import asyncio
import json

import asyncio_redis
from Adafruit_MotorHAT import Adafruit_MotorHAT as MotorHAT
from aiohttp import web

import motors
from config import THE_HEADS_EVENTS, Config
from const import DEFAULT_CONSUL_ENDPOINT
from consul_config import ConsulBackend, ConfigError

STEPPERS_PORT = 8080
NUM_STEPS = 200
DEFAULT_SPEED = 50
directions = {1: MotorHAT.FORWARD, -1: MotorHAT.BACKWARD}
_DEFAULT_REDIS = "127.0.0.1:6379"


class Stepper:
    def __init__(self, cfg, redis: asyncio_redis.Connection):
        self._pos = 0
        self._target = 0
        self._motor = motors.setup()
        self._speed = DEFAULT_SPEED
        self.queue = asyncio.Queue()
        self.cfg = cfg
        self.redis = redis

    @property
    def pos(self) -> int:
        return self._pos

    def zero(self):
        self._pos = 0
        self._target = 0

    def set_target(self, target: int):
        self._target = target

    def set_speed(self, speed: float):
        self._speed = speed

    def current_rotation(self) -> float:
        return self._pos / NUM_STEPS * 360.0

    async def seek(self):
        while True:
            options = [
                ((self._target - self._pos) % NUM_STEPS, 1),
                ((self._pos - self._target) % NUM_STEPS, -1),
            ]

            steps, direction = min(options)

            if steps > 0:
                self._pos += direction
                self._pos %= NUM_STEPS
                self.queue.put_nowait(self._pos)
                self._motor.oneStep(directions[direction], MotorHAT.DOUBLE)

            await asyncio.sleep(1.0 / self._speed)

    async def redis_publisher(self):
        while True:
            pos = await self.queue.get()
            msg = {
                "type": "head-positioned",
                "installation": self.cfg['installation'],
                "data": {
                    "headName": self.cfg['head'],
                    "stepPosition": pos,
                    "rotation": self.current_rotation(),
                }
            }
            await self.redis.publish(THE_HEADS_EVENTS, json.dumps(msg))


def adjust_position(request, speed, target):
    speed = float(speed) if speed else None
    stepper = request.app['stepper']
    stepper.set_target(target)
    if speed is not None:
        stepper.set_speed(speed)
    result = json.dumps({"result": "ok"})
    return web.Response(text=result + "\n", content_type="application/json")


def position(request):
    target = int(request.match_info.get('target'))
    speed = request.query.get("speed")

    return adjust_position(request, speed, target)


def rotation(request):
    theta = float(request.match_info.get('theta'))
    speed = request.query.get("speed")

    target = int(round(theta / 360.0 * NUM_STEPS))

    return adjust_position(request, speed, target)


async def zero(request):
    stepper = request.app['stepper']
    stepper.zero()

    result = json.dumps({"result": "ok"})
    return web.Response(text=result + "\n", content_type="application/json")


async def setup(app: web.Application, args, loop):
    consul_backend = ConsulBackend(args.endpoint)

    # TODO: this is going to read "installation", which doesn't fit the new paradigm
    cfg = await Config(consul_backend).setup(args.instance)

    port = args.port
    if port is None:
        # TODO: should this be using the agent endpoints?
        resp, text = await consul_backend.get_nodes_for_service("heads", tags=[args.instance])
        assert resp.status == 200
        instances = json.loads(text)
        if len(instances) == 0:
            raise ConfigError("No head service defined for {}".format(args.instance))

        if len(instances) > 1:
            raise ConfigError("Multiple head services defined for {}".format(args.instance))

        print(instances[0])
        port = instances[0]['ServicePort']

    redis_server = _DEFAULT_REDIS  # TODO

    redis_host, redis_port_str = redis_server.split(":")
    redis_port = int(redis_port_str)

    print("connecting to redis")
    redis_connection = await asyncio_redis.Connection.create(host=redis_host, port=redis_port)
    print("connected to redis")

    stepper = Stepper(cfg, redis_connection)
    asyncio.ensure_future(stepper.redis_publisher())

    app['stepper'] = stepper

    asyncio.ensure_future(stepper.seek(), loop=loop)

    result = dict(
        endpoint=args.endpoint,
        installation=cfg.installation,
        redis_server=redis_server,
        head=args.instance,
        port=port,
    )

    return result


async def home(request):
    cfg = request.app['cfg']
    stepper = request.app['stepper']
    result = 'This is head "{}"\nPosition is {}'.format(cfg['head'], stepper.pos)
    return web.Response(text=result)


def main():
    parser = argparse.ArgumentParser(description='Process some integers.')

    parser.add_argument('instance', type=str,
                        help='Instance name for this head')

    parser.add_argument('--port', type=int, default=None,
                        help='Port override')

    parser.add_argument('--endpoint', type=str, default="http://127.0.0.1:8500",
                        help='Config service endpoint')

    args = parser.parse_args()

    loop = asyncio.get_event_loop()
    app = web.Application()
    cfg = loop.run_until_complete(setup(app, args, loop))

    app['cfg'] = cfg

    app.add_routes([
        web.get("/", home),
        web.get("/position/{target}", position),
        web.get("/rotation/{theta}", rotation),
        web.get("/zero", zero),
    ])
    web.run_app(app, port=cfg['port'])


if __name__ == '__main__':
    main()
