package erlang

import (
	"encoding/base64"
	"encoding/json"
	"github.com/halturin/ergonode"
	"github.com/halturin/ergonode/etf"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	clientKafka "kafka_producer/pkg/kafka"
)

type State struct {
	KafkaClient *clientKafka.KafkaClient
}

// GenServer implementation structure
type goGenServ struct {
	ergonode.GenServer
	ServerName    string
	kafkaSettings clientKafka.KafkaSettings
	completeChan  chan bool
}

func NewGoGenServ(ServerName string, settings clientKafka.KafkaSettings) *goGenServ {
	return &goGenServ{
		ServerName:    ServerName,
		kafkaSettings: settings,
	}
}

// Init initializes process state using arbitrary arguments
func (gs *goGenServ) Init(args ...interface{}) (state interface{}) {

	log.Info().Msgf("goGenServ Init (%s)", gs.ServerName)

	client, err := clientKafka.NewKafkaWriter(gs.kafkaSettings)
	if err != nil {
		log.Error().Err(err).Msg("Can not connected kafka client")
		panic(err)
	}

	// Self-registration with name gen_server_name
	gs.Node.Register(etf.Atom(gs.ServerName), gs.Self)

	// Store first argument as channel
	gs.completeChan = args[0].(chan bool)

	return State{KafkaClient: client}
}

// HandleCast serves incoming messages sending via gen_server:cast
func (gs *goGenServ) HandleCast(message *etf.Term, state interface{}) (code int, stateout interface{}) {
	return
}

// HandleCall serves incoming messages sending via gen_server:call
func (gs *goGenServ) HandleCall(from *etf.Tuple, message *etf.Term, state interface{}) (code int, reply *etf.Term, stateout interface{}) {
	inState := state.(State)
	stateout = state
	code = 1
	replyTerm := etf.Term(etf.Tuple{etf.Atom("error"), etf.Atom("unknown_request")})
	reply = &replyTerm

	switch req := (*message).(type) {
	case etf.Atom:
		switch string(req) {
		case "pid":
			replyTerm = etf.Term(etf.Pid(gs.Self))
			return
		}

	case etf.Tuple:
		if len(req) == 4 {
			args := req[2].(string)
			requestID := ""

			if str, ok := req[3].(string); ok {
				requestID = str
			}

			var kafkaMessage clientKafka.KafkaMessage

			err := json.Unmarshal([]byte(args), &kafkaMessage)
			if err != nil {
				log.Error().Err(err)
			}

			log.Info().Msgf("request_id: %s message: %+v", requestID, kafkaMessage)

			if len(kafkaMessage.Topic) == 0 {
				replyTerm = etf.Term(etf.Tuple{etf.Atom("error"), "No valid topic"})
				return
			}

			kafkaClient := inState.KafkaClient
			//defer kafkaClient.Client.Close()

			data, err := base64.StdEncoding.DecodeString(kafkaMessage.Message)
			if err != nil {
				log.Error().Err(err).Msg("Invalid base64 string")
				replyTerm = etf.Term(etf.Tuple{etf.Atom("error"), "Invalid base64 string"})
				return
			}

			msg := kafka.Message{
				Value: data,
			}
			_, err = kafkaClient.Client.WriteMessages(msg)
			if err != nil {
				log.Error().Err(err).Msg("Error send message Kafka client")
				replyTerm = etf.Term(etf.Tuple{etf.Atom("error"), "Error send message Kafka client"})
				return
			}

			if err == nil {
				replyTerm = etf.Term(etf.Atom("ok"))
			}
		}
	}
	return
}

// HandleInfo serves all another incoming messages (Pid ! message)
func (gs *goGenServ) HandleInfo(message *etf.Term, state interface{}) (code int, stateout interface{}) {
	log.Info().Msgf("HandleInfo: %#v\n", *message)
	stateout = state
	code = 0
	return
}

// Terminate called when process died
func (gs *goGenServ) Terminate(reason int, state interface{}) {
	log.Info().Msgf("Terminate: %#v\n", reason)
}
