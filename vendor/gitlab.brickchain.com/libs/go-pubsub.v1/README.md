# PubSub library for Brickchain
Abstraction layer and helper for PubSub systems  
The PubSubInterface interface has three methods:
* ```Publish(topic string, docId string) error```, Publish a message to a topic
* ```Subscribe(topic string) (*Subscriber, error)```, Start a subscriber on a topic and return a Subscriber
* ```DeleteTopic(topic string) error```, Clean up a topic

The Subscriber interface has two methods:
* ```Pull(timeout time.Duration) (string, int)```, Get a message. The integer returned is 0 for success or 1 for timeout
* ```Stop()```, Stop the subscriber

## Publisher
```go
import (
    "gitlab.brickchain.com/brickchain/pubsub"
)

func Publish(msg string) {
	p, err := pubsub.NewGCloudPubSub("project-id", "/path/to/credentials.json")
	if err != nil {
		panic(err)
	}
	
	err = p.Publish("brickchain/documentType", "documentId")
	if err != nil {
	    panic(err)
	}
}

```

## Subscriber
```go
import (
    "fmt"
    "gitlab.brickchain.com/brickchain/pubsub"
)

func Subscriber() {
    	p, err := pubsub.NewGCloudPubSub("project-id", "/path/to/credentials.json")
    	if err != nil {
    		panic(err)
    	}
    	
        sub, err := p.Subscribe("subscriber_group_name", "brickchain/documentType")
        if err != nil {
            t.Error(err)
        }
        
        for i := 0; i < 10; i++ {
            msg, ok := sub.Pull(10)
            if ok == TIMEOUT {
                fmt.Println("Pull timed out")
            }
            fmt.Println("Received message:", msg)
        }
}
```

## Start the Google Cloud PubSub emulator
```bash
gcloud beta emulators pubsub start --host-port=localhost:9111
```
Then run ```export PUBSUB_EMULATOR_HOST=localhost:9111``` in the shell where you will run the tests.  
