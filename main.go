package main

import (
	"github.com/halturin/ergonode"
	"gopkg.in/alecthomas/kingpin.v2"
	"runtime"

	"github.com/rs/zerolog/log"
	"kafka_producer/pkg/erlang"
	clientKafka "kafka_producer/pkg/kafka"
)

var (
	listenAddressKafka = kingpin.Flag("kafka.listen-address", "Address to listen on for kafka client.").Default("localhost:9092").String()
	networkKafka       = kingpin.Flag("kafka.network", "Network for kafka client.").Default("tcp").String()
	topicKafka         = kingpin.Flag("kafka.topic", "Topic for kafka client.").Default("").String()
	partitionKafka     = kingpin.Flag("kafka.partition", "Partition for kafka client.").Default("0").Int()

	erlangCookie = kingpin.Flag("erlang.cookie", "cookie for interaction with erlang cluster").Default("123").String()

	//serverGenName = kingpin.Flag("server.gen_name", "gen_server name").Default("gen_server_name").String()
	serverNodeName = kingpin.Flag("server.node_name", "node name").Default("examplenode@127.0.0.1").String()
	empdPort       = kingpin.Flag("empd.port", "Init empd port").Default("15151").Uint16()
)

func main() {
	defer func() {
		if e := recover(); e != nil {

			if e != nil {
				if _, ok := e.(runtime.Error); ok {
					panic(e)
				}
				log.Error().Msgf("Error recover: %s", e)
			}
		}
	}()

	kingpin.Parse()

	log.Info().Msgf("Start producer kafka %s", *listenAddressKafka)

	// Initialize new node with given name and cookie
	n := ergonode.Create(*serverNodeName, *erlangCookie, *empdPort)
	log.Info().Msg("Started erlang node")

	// Create channel to receive message when main process should be stopped
	completeChan := make(chan bool)

	settingsKafka := clientKafka.KafkaSettings{
		ListenAddress: *listenAddressKafka,
		Topic:         *topicKafka,
		Network:       *networkKafka,
		Partition:     *partitionKafka,
	}

	gs := erlang.NewGoGenServ(*serverNodeName, settingsKafka)
	// Spawn process with one arguments
	n.Spawn(gs, completeChan)
	log.Info().Msg("Spawned gen server process")

	// Wait to stop
	<-completeChan

	return
}
