import asyncio
import json
import sys
import time

import asyncio_redis
from Adafruit_MotorHAT import Adafruit_MotorHAT as MotorHAT
from aiohttp import web

import motors
from const import DEFAULT_CONSUL_ENDPOINT
from consul_config import ConsulBackend
from config import THE_HEADS_EVENTS, Config, get_redis

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
                    "position": pos,
                }
            }
            await self.redis.publish(THE_HEADS_EVENTS, json.dumps(msg))


def position(request):
    stepper = request.app['stepper']
    target = int(request.match_info.get('target'))
    speed = request.query.get("speed")
    speed = float(speed) if speed else None

    stepper.set_target(target)
    if speed is not None:
        stepper.set_speed(speed)

    result = json.dumps({"result": "ok"})
    return web.Response(text=result + "\n", content_type="application/json")


async def zero(request):
    stepper = request.app['stepper']
    stepper.zero()

    result = json.dumps({"result": "ok"})
    return web.Response(text=result + "\n", content_type="application/json")


def console_fun():
    stepper = motors.setup()
    steps = int(sys.argv[1])

    direction = MotorHAT.FORWARD if steps >= 0 else MotorHAT.BACKWARD
    steps = abs(steps)

    while True:
        for i in range(steps):
            stepper.oneStep(MotorHAT.FORWARD, MotorHAT.DOUBLE)
            # stepper.oneStep(MotorHAT.BACKWARD, MotorHAT.SINGLE)
            time.sleep(0.01)

        for i in range(steps):
            stepper.oneStep(MotorHAT.BACKWARD, MotorHAT.DOUBLE)
            # stepper.oneStep(MotorHAT.BACKWARD, MotorHAT.SINGLE)
            time.sleep(0.01)


async def get_config(endpoint: str):
    cfg = await Config(ConsulBackend(endpoint)).setup()

    redis_server = _DEFAULT_REDIS  # TODO

    head = await cfg.get_config_str("/the-heads/installation/{installation}/heads/{hostname}")

    return dict(
        endpoint=endpoint,
        installation=cfg.installation,
        redis_server=redis_server,
        head=head,
    )


async def setup(app: web.Application, loop):
    cfg = await get_config(DEFAULT_CONSUL_ENDPOINT)

    redis_host, redis_port_str = cfg['redis_server'].split(":")
    redis_port = int(redis_port_str)

    redis_connection = await asyncio_redis.Connection.create(host=redis_host, port=redis_port)

    stepper = Stepper(cfg, redis_connection)
    asyncio.ensure_future(stepper.redis_publisher())

    app['stepper'] = stepper

    asyncio.ensure_future(stepper.seek(), loop=loop)
    return cfg


async def home(request):
    cfg = request.app['cfg']
    stepper = request.app['stepper']
    result = 'This is head "{}"\nPosition is {}'.format(cfg['head'], stepper.pos)
    return web.Response(text=result)


def main():
    loop = asyncio.get_event_loop()
    app = web.Application()
    cfg = loop.run_until_complete(setup(app, loop))

    app['cfg'] = cfg

    app.add_routes([
        web.get("/", home),
        web.get("/position/{target}", position),
        web.get("/zero", zero),
    ])
    web.run_app(app, port=STEPPERS_PORT)


if __name__ == '__main__':
    main()
