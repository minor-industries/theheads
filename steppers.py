import sys
import time

from Adafruit_MotorHAT import Adafruit_MotorHAT as MotorHAT
from bottle import run, template, Bottle, request

from motors import setup

app = Bottle()
stepper = setup()
pos = 0

NUM_STEPS = 200

DEFAULT_SPEED = 50

directions = {1: MotorHAT.FORWARD, -1: MotorHAT.BACKWARD}


@app.route('/forward/<steps>')
def index(steps):
    global pos

    speed = float(request.params.get("speed") or DEFAULT_SPEED)

    steps = int(steps)
    direction = 1 if steps >= 0 else -1
    steps = abs(steps)
    for i in range(steps):
        pos += direction
        stepper.oneStep(directions[direction], MotorHAT.DOUBLE)
        time.sleep(1 / speed)
    return template('Hello {{steps}}\n', steps=steps)


@app.route('/position/<target>')
def index(target):
    global pos

    target = int(target)

    speed = float(request.params.get("speed") or DEFAULT_SPEED)
    print(speed)

    options = [
        ((target - pos) % NUM_STEPS, 1),
        ((pos - target) % NUM_STEPS, -1),
    ]

    steps, direction = min(options)

    for i in range(steps):
        pos += direction
        stepper.oneStep(directions[direction], MotorHAT.DOUBLE)
        time.sleep(1 / speed)
    return template('at {{pos}}\n', pos=pos)


@app.route('/zero')
def index():
    global pos
    pos = 0


def console_fun():
    stepper = setup()
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


def main():
    run(app, host='0.0.0.0', port=8080, reloader=True)


if __name__ == '__main__':
    main()