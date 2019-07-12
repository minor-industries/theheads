package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var upgrader = websocket.Upgrader{}

type HeadEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type MotionDetected struct {
	CameraName string  `json:"cameraName"`
	Position   float32 `json:"position"`
}

type HeadPositioned struct {
	HeadName     string  `json:"headName"`
	StepPosition float32 `json:"stepPosition"`
	Rotation     float32 `json:"rotation"`
}

func runRedis(broker *Broker, redisServer string) {
	log.Println("Connecting to redis @ ", redisServer)
	redisClient, err := redis.Dial("tcp", redisServer)
	if err != nil {
		log.WithError(err).Error("Error connecting to redis")
		panic(err)
	}
	log.Println("Connected to redis @ ", redisServer)

	psc := redis.PubSubConn{Conn: redisClient}

	if err := psc.Subscribe("the-heads-events"); err != nil {
		panic(err)
	}

	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			log.Println("message:", v.Channel, string(v.Data))
			event := HeadEvent{}
			err := json.Unmarshal(v.Data, &event)
			if err != nil {
				panic(err)
			}
			switch event.Type {
			case "head-positioned":
				msg := HeadPositioned{}
				err = json.Unmarshal(event.Data, &msg)
				if err != nil {
					panic(err)
				}
				broker.Publish(msg)
			}
		case redis.Subscription:
			fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			panic(v)
		}
	}
}

func main() {
	broker := NewBroker()
	go broker.Start()

	redisServers := []string{"127.0.0.1:6379"}

	for _, redis := range redisServers {
		go runRedis(broker, redis)
	}

	addr := ":7071"

	r := gin.New()
	r.Use(
		gin.LoggerWithWriter(gin.DefaultWriter, "/metrics", "/health"),
		gin.Recovery(),
	)

	var sceneJson interface{}
	err := json.Unmarshal([]byte(scene), &sceneJson)
	if err != nil {
		panic(err)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"result": "ok",
		})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.StaticFile("/", "./templates/boss.html")

	//web.get("/build/{filename}", frontend_handler("boss-ui/build")),
	//web.get("/build/json/{filename}", frontend_handler("boss-ui/build/json")),
	//web.get("/build/media/{filename}", frontend_handler("boss-ui/build/media")),
	//web.get("/build/js/{filename}", frontend_handler("boss-ui/build/js")),
	//web.get("/static/js/{filename}", frontend_handler("boss-ui/build/static/js")),
	//web.get("/static/css/{filename}", frontend_handler("boss-ui/build/static/css")),

	/*
				# deprecated, use don't use above instead
		        web.get('/installation/{installation}/scene.json', installation_handler),
		        web.get('/installation/{installation}/{name}.html', html_handler),
		        web.get('/installation/{installation}/{name}.js', static_text_handler("js")),
		        web.get('/installation/{installation}/{name}.png', static_binary_handler("png")),

	*/

	r.GET("/installation/:installation/scene.json", func(c *gin.Context) {
		c.JSON(200, sceneJson)
	})

	r.Static("/build", "./boss-ui/build")
	//r.Static("/build/json", "./boss-ui/build")
	r.Static("/static", "boss-ui/build/static")

	r.GET("/ws", gin.WrapF(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Failed to set websocket upgrade: %+v", err)
			return
		}

		msgs := broker.Subscribe()

		for {
			//type_, msg, err := conn.ReadMessage()
			//if err != nil {
			//	break
			//}
			//conn.WriteMessage(type_, msg)
			//log.Println("ws", type_, string(msg))

			for i := range msgs {
				m := i.(HeadPositioned)

				data, err := json.Marshal(m)
				if err != nil {
					panic(err)
				}

				events := []HeadEvent{{
					Type: "head-positioned",
					Data: data,
				}}

				payload, err := json.Marshal(events)
				if err != nil {
					panic(err)
				}

				conn.WriteMessage(1, payload)
			}
		}
	}))

	go func() {
		r.Run(addr)
	}()

	select {}
}
