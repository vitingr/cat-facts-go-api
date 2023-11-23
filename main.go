package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BSON = Binary Javascript Object Notation

// Structure do Servidor
type Server struct {
	client *mongo.Client
}

// Função de Inicializaçao do Servidor Web Http
func NewServer(c *mongo.Client) *Server {
	return &Server {
		client: c,
	}
}

func (s *Server) handleGetAllFacts(w http.ResponseWriter, r*http.Request) {
	// a letra antes do ponteiro (*) indica um rename para a funçõa, tipo s = *Server
	call := s.client.Database("catfact").Collection("facts")

	query := bson.M{}

	// Método FIND para buscar todos dados do MongoDB
	cursor, err := call.Find(context.TODO(), query)
	if err != nil {
		log.Fatal(err)
	}

	// Transformar os resultados em JSON
	results := []bson.M{}

	// Check for errors in the conversion
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
	}

	// Métodos igual do Next, StatusOK = {status: 200}
	// Content-type igual do Fetch
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")

	// Transformar os dados retornados do banco de dados em JSON
	json.NewEncoder(w).Encode(results)
}

// Structure da Response da API
type CatFactWorker struct {
	client *mongo.Client
}

func NewCatFactWorker(c *mongo.Client) *CatFactWorker {
	return &CatFactWorker {
		client: c,
	}
}

// Função de Fetch na URL e INSERT da response no banco de dados
func (cfw *CatFactWorker) start() error {
	call := cfw.client.Database("catfact").Collection("facts")
	ticker := time.NewTicker(2 * time.Second)

	// Laço de Repetição pelo ticker
	for {
		// Fetch na URL
		resp, err := http.Get("https://catfact.ninja/fact")
		if err != nil {
			return err
		}
		var catFact bson.M // map[string]any //map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&catFact); err != nil {
			return err
		}
   
		// Método de Insert no Mongo
		_, err = call.InsertOne(context.TODO(), catFact)
		if err != nil {
			return err
		}

		<-ticker.C
	}
}

// Função Principal
func main() {
	// Configurações da conexão do banco de dados
	uri := "mongodb+srv://vitorgabrielsbo1460:xXzMW9c0UljiQykk@aprendendo.cmdbthe.mongodb.net/golang?retryWrites=true&w=majority"
	if uri == "" {
		log.Fatal("ERROR! Connection with MongoDB failed...")
	}
	// Tratamento de Erros
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	// Método de conexão
	worker := NewCatFactWorker(client)
	go worker.start()

	// Inicialização das funções do servidor
	server := NewServer(client)
	http.HandleFunc("/facts", server.handleGetAllFacts)

	// Listen e definição da porta padrão
	http.ListenAndServe(":3000", nil)
}
