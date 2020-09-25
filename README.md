[![GoDoc](https://godoc.org/github.com/LdDl/fiber-long-poll?status.svg)](https://godoc.org/github.com/LdDl/fiber-long-poll)
[![Go Report Card](https://goreportcard.com/badge/github.com/LdDl/fiber-long-poll)](https://goreportcard.com/report/github.com/LdDl/fiber-long-poll)

# Long polling library for [Fiber](https://github.com/gofiber/fiber) web-framework

Golang long polling library for [fasthttp](https://github.com/valyala/fasthttp)-based web framework called [Fiber](https://github.com/gofiber/fiber).

Makes web pub-sub easy via an HTTP long-poll server.

## Table of Contents

- [About](#about)
- [Usage](#usage)
- [Issues](#issues)
- [License](#license)

## About
This library is just a port of existing library for long polling https://github.com/jcuga/golongpoll, but for Fiber ecosystem.
You can read about it here https://github.com/jcuga/golongpoll#table-of-contents.

## Usage
Here is example code: [click!](example)

Golang server side:
```go
package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	lp "github.com/LdDl/fiber-long-poll/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/valyala/fasthttp"
)

var (
	userID = "my_pretty_uuid"
)

func main() {

	config := fiber.Config{
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			log.Println(err)
			return ctx.Status(fasthttp.StatusInternalServerError).JSON(errors.New("panic error"))
		},
		IdleTimeout: 10 * time.Second,
	}
	allCors := cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "Origin, Authorization, Content-Type, Content-Length, Accept, Accept-Encoding, X-HttpRequest",
		AllowMethods:     "GET, POST, PUT, DELETE",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: true,
		MaxAge:           5600,
	})

	server := fiber.New(config)
	server.Use(allCors)

	manager, err := lp.StartLongpoll(lp.Options{
		LoggingEnabled:                 false,
		MaxLongpollTimeoutSeconds:      120,
		MaxEventBufferSize:             100,
		EventTimeToLiveSeconds:         60 * 2,
		DeleteEventAfterFirstRetrieval: false,
	})
	if err != nil {
		log.Printf("Failed to create manager: %q", err)
		return
	}
	defer manager.Shutdown()

	go generatingMessages(manager)
	server.Get("/unread_messages", GetMessages(manager))

	err = server.Listen(":8080")
	if err != nil {
		fmt.Printf("Can't start server due the error: %s\n", err.Error())
	}
}

// GetMessages Long polling request
func GetMessages(manager *lp.LongpollManager) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		ctx.Context().PostArgs().Set("timeout", "10")
		ctx.Context().PostArgs().Set("category", fmt.Sprintf("unread_messages_for_%s", userID))
		return manager.SubscriptionHandler(ctx)
	}
}

// generatingMessages Generate some messages
func generatingMessages(manager *lp.LongpollManager) {
	i := 0
	for {
		manager.Publish(fmt.Sprintf("unread_messages_for_%s", userID), fmt.Sprintf("Number: %d", i))
		i++
		time.Sleep(3 * time.Second)
	}
}
```

JavaScript client side:
```js

<html>
    <head>
        <title>Fiber long polling example</title>
    </head>
    <body>
        <ul id="unred-messages"></ul>
        <script src="http://code.jquery.com/jquery-1.11.3.min.js"></script>
        <script>
            if(typeof window.console == 'undefined') { window.console = {log: function (msg) {} }; }
            var sinceTime = (new Date(Date.now())).getTime();
            (function poll() {
                var timeout = 45;  // in seconds
                var optionalSince = "";
                if (sinceTime) {
                    optionalSince = "&since_time=" + sinceTime;
                }
                var pollUrl = `http://localhost:8080/unread_messages`;
                // how long to wait before starting next longpoll request in each case:
                var successDelay = 10;  // 10 ms
                var errorDelay = 3000;  // 3 sec
                $.ajax({ url: pollUrl,
                    success: function(data) {
                        if (data && data.events && data.events.length > 0) {
                            // got events, process them
                            // NOTE: these events are in chronological order (oldest first)
                            for (var i = 0; i < data.events.length; i++) {
                                // Display event
                                var event = data.events[i];
                                $("#unred-messages").append("<li>" + JSON.stringify(event.data) + " at " + (new Date(event.timestamp).toLocaleTimeString()) +  "</li>")
                                // Update sinceTime to only request events that occurred after this one.
                                sinceTime = event.timestamp;
                            }
                            console.log(data.events);
                            // success!  start next longpoll
                            setTimeout(poll, successDelay);
                            return;
                        }
                        if (data && data.timeout) {
                            console.log("No events, checking again.");
                            // no events within timeout window, start another longpoll:
                            setTimeout(poll, successDelay);
                            return;
                        }
                        if (data && data.error) {
                            console.log("Error response: " + data.error);
                            console.log("Trying again shortly...")
                            setTimeout(poll, errorDelay);
                            return;
                        }
                        // We should have gotten one of the above 3 cases:
                        // either nonempty event data, a timeout, or an error.
                        console.log("Didn't get expected event data, try again shortly...");
                        setTimeout(poll, errorDelay);
                    }, dataType: "json",
                error: function (data) {
                    console.log("Error in ajax request--trying again shortly...");
                    setTimeout(poll, errorDelay);  // 3s
                }
                });
            })();
        </script>
    </body>
</html>
```

## Support
If you have troubles or questions please [open an issue](https://github.com/LdDl/fiber-long-poll/issues/new).

## License
You can check it [here](LICENSE.md)