package main

import (
	"log"

	rovers "github.com/mcarmonaa/mini-rovers"
	"gopkg.in/src-d/go-queue.v1"
	_ "gopkg.in/src-d/go-queue.v1/amqp"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Queue  string `long:"queue" env:"QUEUE_NAME" default:"mini-rovers" description:"queue name"`
	Broker string `long:"broker" env:"BROKER_URI" default:"amqp://localhost:5672" description:"broker service URI"`
	Token  string `short:"t" long:"token" env:"GH_TOKEN" description:"github authentication token "`
	List   struct {
		Path string `positional-arg-name:"list" description:"path to a file containing a list of githug organizations(one per line)" required:"true"`
	} `positional-args:"true"`
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal()
	}

	b, err := queue.NewBroker(opts.Broker)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := b.Close(); err != nil {
			log.Println(err)
		}
	}()

	q, err := b.Queue(opts.Queue)
	if err != nil {
		log.Fatal(err)
	}

	iter, err := rovers.NewOrganizationIterator(opts.List.Path)
	if err != nil {
		log.Fatal(err)
	}

	provider := rovers.NewProvider(
		rovers.EnqueueMention(q),
		iter,
		opts.Token,
	)

	if err := provider.Start(); err != nil {
		log.Fatal(err)
	}
}
