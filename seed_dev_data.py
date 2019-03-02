import asyncio

import yaml

from const import DEFAULT_CONSUL_ENDPOINT
from consul_config import ConsulBackend

INSTALLATION = "dev"


async def main(inst_name: str):
    consul_backend = ConsulBackend(DEFAULT_CONSUL_ENDPOINT)

    async def put(key: str, value: bytes):
        print(key, value)
        resp, _ = await consul_backend.put(key.encode(), value)
        assert resp.status == 200

    async def setup_services():
        with open('seed_data/{}.yaml'.format(inst_name), "r") as fp:
            inst_data = yaml.safe_load(fp)

        for stand in inst_data['stands']:
            for camera in stand.get('cameras', []):
                await put(
                    "/the-heads/installation/{}/cameras/{}.yaml".format(inst_name, camera['name']),
                    yaml.dump(camera, encoding='utf-8'),
                )

            for head in stand['heads']:
                await put(
                    "/the-heads/installation/{}/heads/{}.yaml".format(inst_name, head['name']),
                    yaml.dump(head, encoding='utf-8'),
                )

            stand['cameras'] = [x['name'] for x in stand.get('cameras', [])]
            stand['heads'] = [x['name'] for x in stand['heads']]
            key = "/the-heads/installation/{}/stands/{}.yaml".format(inst_name, stand['name'])
            value = yaml.dump(stand, encoding='utf-8')
            await put(key, value)

    async def setup_instances():
        for i in range(11):
            name = "vhead-{:02}".format(i)
            await consul_backend.register_service_with_agent("heads", 18080+i, ID=name, tags=[name, "frontend"])
            await put("/the-heads/assignment/{}".format(name), inst_name.encode())

        # redis
        await consul_backend.register_service_with_agent("redis", 6379)

        # boss
        await consul_backend.register_service_with_agent("boss", 8081, ID="boss-00", tags=["boss-00", "frontend"])
        await put("/the-heads/assignment/boss-00", inst_name.encode())

        # consul-fe
        await consul_backend.register_service_with_agent("consul-fe", 8500, tags=["frontend"])

        # home
        await consul_backend.register_service_with_agent("home", 8500, tags=["frontend"])

    await setup_services()
    await setup_instances()


if __name__ == '__main__':
    loop = asyncio.get_event_loop()
    loop.run_until_complete(main(INSTALLATION))
