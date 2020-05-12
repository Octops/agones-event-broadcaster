# [Alpha] GameServer Events Broadcaster

Broadcast Agones GameServer reconcile events using a message queueing service (or any other implementation of the Broker).

### Agones
> An open source, batteries-included, multiplayer dedicated game server scaling and orchestration platform that can run anywhere Kubernetes can run.

You can find a great documentation on https://agones.dev/site/

### Use Cases
The most common use case is for folks who want to extract information about their Agones GameServers running within Kubernetes.
Using the broadcaster, you can have one single point of extraction and publish it to different destinations.

Currently, the project supports Google Cloud Pub/Sub. However, this can be extended to any kind of backend.
Your implementation of the broker can publish the event to databases, message queueing services, request REST APIs, communicate with gRPC services, etc. There is virtually no restrictions.    

You don't need/use/require a message queuing service? We got you covered. Implement your own broker and plug it to the broadcaster.
Possible, but not limited, ideas for brokers are:
- MongoDB
- Elasticsearch
- MySQL
- Kafka
- Cassandra
- CockroachDB
- REST Endpoint
- Remote storage (GCS, S3)
- Logging systems
- RabbitMQ
- gRPC Endpoint
- BigData
- Google Cloud Datastore
- Game server backend
- Match maker
- ...

## FAQ
### What does GameServer event mean?
For the broadcaster, an event is some sort of information that reflects a GameServer state running within a Kubernetes cluster in a particular moment in time.

### How/Where events can be emitted?
The source of events can vary depending on the action which triggered those. For instance, when a GameServer is deployed it might change its state many times.
From the port allocation to the ready state, information will be added and updated. I.e.: Address and port, status and labels.

### Who is responsible for watching events from the Kubernetes API?
The broadcaster implements the [Kubernetes Controller Pattern](https://kubernetes.io/docs/concepts/architecture/controller/#controller-pattern). It tracks events for resources of type GameServer.

For every single time the state of a resource of type GameServer has changed, the broadcaster will be notified. Therefore, it will handle the event and publish a message using a Broker.
Currently, the controller watches GameServers deployed on any namespace. It may change in the future if that becomes a performance issue.

### What kind of events are tracked?
The broadcaster watches for Add, Update and Delete events.
- Add: When a new GameServer is deployed by Agones
- Update: When the GameServer state changes by changing any information from its specification.
- Delete: When the GameServer is deleted from the Kubernetes cluster

### What does the event message content look like?
The current version of the broadcaster sends the entire Agones GameServer state representation as an encoded json. Additionally, some headers holding information about event type and event source.
In the future there may be a event middleware/parser that could extract pieces of information and make the message of customisable. 

## How to run the broadcaster?

Requirements
 - Linux or OSX
 - Kubernetes Cluster 1.14 (Supported by Agones 1.15)
 - Agones 1.15


## Supported Brokers
Below you can find a list a supported brokers that can be used for publishing messages.

### Logging - Stdout [Development purpose]
Only outputs the content of the message to the application `stdout`.

It can be used for local development of when another type of broker is not available

Output:
```bash
{"message":"{\"header\":{\"headers\":{\"event_type\":\"gameserver.events.added\"}},\"message\":{\"kind\":\"GameServer\",\"apiVersion\":\"agones.dev/v1\",\"metadata\":{\"name\":\"simple-udp-agones\",\"namespace\":\"default\",\"selfLink\":\"/apis/agones.dev/v1/namespaces/default/gameservers/simple-udp-agones\",\"uid\":\"2762bdb9-9387-11ea-ab97-0242ac110002\",\"resourceVersion\":\"827719\",\"generation\":6,\"creationTimestamp\":\"2020-05-11T12:58:47Z\",\"annotations\":{\"agones.dev/ready-container-id\":\"containerd://ca32c651fa5f34efbd263145a9423123616e39ab3e3ddc83358e594cce442236\",\"agones.dev/sdk-version\":\"1.5.0\",\"kubectl.kubernetes.io/last-applied-configuration\":\"{\\\"apiVersion\\\":\\\"agones.dev/v1\\\",\\\"kind\\\":\\\"GameServer\\\",\\\"metadata\\\":{\\\"annotations\\\":{},\\\"name\\\":\\\"simple-udp-agones\\\",\\\"namespace\\\":\\\"default\\\"},\\\"spec\\\":{\\\"ports\\\":[{\\\"containerPort\\\":7654,\\\"name\\\":\\\"default\\\",\\\"portPolicy\\\":\\\"Dynamic\\\"}],\\\"template\\\":{\\\"spec\\\":{\\\"containers\\\":[{\\\"image\\\":\\\"gcr.io/agones-images/udp-server:0.18\\\",\\\"name\\\":\\\"simple-udp\\\",\\\"resources\\\":{\\\"limits\\\":{\\\"cpu\\\":\\\"20m\\\",\\\"memory\\\":\\\"32Mi\\\"},\\\"requests\\\":{\\\"cpu\\\":\\\"20m\\\",\\\"memory\\\":\\\"32Mi\\\"}}}]}}}}\\n\"},\"finalizers\":[\"agones.dev\"]},\"spec\":{\"container\":\"simple-udp\",\"ports\":[{\"name\":\"default\",\"portPolicy\":\"Dynamic\",\"containerPort\":7654,\"hostPort\":7100,\"protocol\":\"UDP\"}],\"health\":{\"periodSeconds\":5,\"failureThreshold\":3,\"initialDelaySeconds\":5},\"scheduling\":\"Packed\",\"sdkServer\":{\"logLevel\":\"Info\",\"grpcPort\":9357,\"httpPort\":9358},\"template\":{\"metadata\":{\"creationTimestamp\":null},\"spec\":{\"containers\":[{\"name\":\"simple-udp\",\"image\":\"gcr.io/agones-images/udp-server:0.18\",\"resources\":{\"limits\":{\"cpu\":\"20m\",\"memory\":\"32Mi\"},\"requests\":{\"cpu\":\"20m\",\"memory\":\"32Mi\"}}}]}}},\"status\":{\"state\":\"Ready\",\"ports\":[{\"name\":\"default\",\"port\":7100}],\"address\":\"172.17.0.2\",\"nodeName\":\"agones-cluster-control-plane\",\"reservedUntil\":null,\"players\":null}}}","severity":"info","time":"2020-05-11T19:40:23.941715+02:00"}
``` 

### Google Cloud Pub/Sub

Publishes messages to [Google Cloud Pub/Sub](https://cloud.google.com/pubsub/docs/overview) topics.

Be aware that using the service may cost you some money. Check https://cloud.google.com/pubsub/pricing for detailed information. If you are just experimenting the project locally, you can use the Google Cloud Pub/Sub emulator https://cloud.google.com/pubsub/docs/emulator.

When publishing a message to Pub/Sub the broker will output the information below.
Output:
```bash
{"broker":"pubsub","message":"message published to topicID:\"gameserver.events.added\" messageID:\"20\"","severity":"info","time":"2020-05-11T19:41:57.607351+02:00"}
```

Requirements:
- Service Account Credentials with `PubSub Editor` role assigned to it. Required for checking if topic exists before publishing.
- Topics created beforehand. Use those topics when creating the broker config.
- Environment variable `PUBSUB_CREDENTIALS`: Json key file path

***Creating the broker***

The topics can be customised by event source (Add, Update, Delete) or be unique for all types of events.

```go
opts := option.WithCredentialsFile(os.Getenv("PUBSUB_CREDENTIALS"))
broker, err := pubsub.NewPubSubBroker(&pubsub.Config{
    ProjectID:       os.Getenv("PUBSUB_PROJECT_ID"),
    // Alternatively, set one topic to publish all types of events.
    // GenericTopicID: "gameserver.events"  
    OnAddTopicID:    "gameserver.events.added", // Use any available topic you you want 
    OnUpdateTopicID: "gameserver.events.updated", // Use any available topic you you want
    OnDeleteTopicID: "gameserver.events.deleted", // Use any available topic you you want
}, opts)
```

Check the [`examples/pubsub/main.go`](examples/pubsub/main.go) file for a complete example.
```bash
$ go run examples/pubsub/main.go 
```

## Development

The steps below provide the instructions for running the broadcaster on your local laptop. We will be using the `stdout ` broker. That means messages will not be published to any remote service. 

Requirements:
- Kubernetes cluster  
- Valid `KUBECONFIG`
- Proper RBAC settings 

```bash 
$ go run main.go --kubeconfig=$KUBECONFIG
```

Running

On another terminal session push the GameServers manifest to the Kubernetes API. You can use [examples/gameservers/agones-udp.yaml](examples/gameservers/agones-udp.yaml) for a simple game server.

Apply the manifest and check the broadcaster output for details.
```bash
# Triggers Add and Update events
$ kubectl apply -f examples/agones-udp.yaml

# Triggers a Delete event
$ kubectl delete -f examples/agones-udp.yaml
```

Tests
```bash
$ make test
```

## Implementing my own broker

As previously mentioned, you can implement your own broker and plug it to the broadcaster. The broker interface is minimal and can be used for different purposes.

The Broker interface:
 
```go
// Broker is the service used by the Broadcaster for publishing events
type Broker interface {
	BuildEnvelope(event events.Event) (*events.Envelope, error)
	SendMessage(envelope *events.Envelope) error
}
```

Example:

```go
// Create an instance of the broker and give it to the broadcaster

// Kafka Broker
import broker ..../brokers/kafka

...

broker := broker.NewKafkaBroker(&broker.Config{})
gsBroadcaster, err := broadcaster.New(clientConf, broker)

...

err := broadCaster.Start()

// MongoDB Broker
import broker ..../brokers/mongodb

...

broker := mongodb.NewMongoDBBroker(&mongodb.Config{})
gsBroadcaster, err := broadcaster.New(clientConf, broker)

...

 err := broadCaster.Start()
```

## Deploying

Build docker image

RBAC, Manifests, 



