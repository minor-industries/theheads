function draw_stand(scene, parent, w, h, stand) {
    var radius = .20;

    stand.heads.forEach(function (head) {
        var g2 = parent.group();

        var circle = g2.circle(0).radius(radius);

        circle.attr({
            fill: '#f06',
            stroke: '#000',
            "stroke-width": 0.020
        });

        g2.line(
            0,
            0,
            radius,
            0
        ).stroke({width: 0.020});

        scene.heads[head.name] = g2;
    });

    stand.cameras.forEach(function (camera) {
        var fov_x = 3;
        var fov_y = fov_x * Math.tan(camera.fov * Math.PI / 180 / 2);

        var g2 = parent.group();

        g2.polygon([0, 0, fov_x, fov_y, fov_x, -fov_y]).attr({
            'fill': "lightblue",
            'fill-opacity': 0.4
        });

        g2.line([0, 0, fov_x, fov_y]).stroke({width: 0.020, color: "darkblue"});
        g2.line([0, 0, fov_x, -fov_y]).stroke({width: 0.020, color: "darkblue"});

        g2.move(camera.pos.x, camera.pos.y);

        scene.cameras[camera.name] = g2;
    });
}


function setup_websocket(ws_port, scene) {
    var url = 'ws://' + window.location.hostname + ":" + ws_port + '/ws';
    console.log(url);
    var ws = new WebSocket(url);
    console.log(ws);

    ws.onmessage = function (event) {
        var msgs = JSON.parse(event.data);
        var info = msgs.map(function (msg) {
            return msg.type;
        }).join(", ");

        console.log("WebSocket message received:", info);

        msgs.forEach(function (msg) {
            if (msg.type === "motion-detected") {
                var cameraName = msg.data.cameraName;
                var camera = scene.cameras[cameraName];
                if (camera === undefined) {
                    console.log("unknown camera: " + cameraName)
                } else {
                    var ray = camera.line(0, 0, 5, 0).stroke({
                        width: 0.020,
                        color: "lightgreen",
                        opacity: 0.40
                    });
                    ray.rotate(msg.data.position, 0, 0);
                    setTimeout(function () {
                        ray.remove();
                    }, 5000);
                }
            } else if (msg.type === "draw") {
                if (msg.data.shape === "line") {
                    var coords = msg.data.coords;
                    var ray = scene.root.line(coords[0], coords[1], coords[2], coords[3]).stroke({
                        width: 0.020,
                        color: "lightgreen",
                        opacity: 0.40
                    });
                    setTimeout(function () {
                        ray.remove();
                    }, 1000);
                }
            } else if (msg.type === "head-positioned") {
                var data = msg.data;
                console.log(data);
                var head = scene.heads[data.headName];
                head.rotate(data.rotation);
            } else if (msg.type === "active") {
                if (msg.data.component === "head") {
                    var extra = msg.data.extra;
                    var head = scene.heads[extra.headName];
                    head.rotate(extra.rotation);
                }
            }
        });
    }
}

function get_json(url, callback) {
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function () {
        if (xhr.readyState === XMLHttpRequest.DONE) {
            var msg = JSON.parse(xhr.responseText);
            callback(msg);
        }
    };
    xhr.open('GET', url, true);
    xhr.send(null);
}


function main(ws_port) {
    var w = 1000;
    var h = 800;

    console.log("here we are");
    var svg = SVG('drawing').size(w, h);
    svg.rect(w, h).attr({fill: "black"});

    var root = svg.group();
    var scale = 33;

    root.move(w / 2, 100);
    root.scale(scale, -scale, 0, 0);

    var img = root.image("abc123/random.png", 20, 40).move(-10, -20).attr({opacity: 1});

    bgReload = function (delay) {
        var name = Math.random() + "/random.png";
        img.load(name);
        setTimeout(function() {
            bgReload(delay);
        }, delay);
    }

    get_json("/installation/dev/scene.json", function (data) {
        root.line(0, 0, 1, 0).stroke({width: 0.040, color: "red"});
        root.line(0, 0, 0, 1).stroke({width: 0.040, color: "lightgreen"});

        var scene = {
            cameras: {},
            heads: {},
            root: root
        };

        data.stands.forEach(function (stand) {
            var parent = root.group();
            parent.dmove(stand.pos.x, stand.pos.y);
            parent.rotate(stand.rot, 0, 0);
            draw_stand(scene, parent, w, h, stand);
        });

        setup_websocket(ws_port, scene);
    });
}