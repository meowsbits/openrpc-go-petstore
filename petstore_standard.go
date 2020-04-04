package petstore

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/alecthomas/jsonschema"
	openrpc_go_document "github.com/etclabscore/openrpc-go-document"
	"github.com/gregdhill/go-openrpc/types"
)

type PetStoreStdService struct {
	pets []*Pet
}

// GetPets returns all the pets the store has.
type None interface{}
type GetPetsArgs None
type GetPetsRes []*Pet

func (s *PetStoreStdService) GetPets(args GetPetsArgs, response *GetPetsRes) error {
	// Returns all pets.
	copy(*response, s.pets)
	return nil
}

// AddPet adds a pet to the store.
type AddPetRes Pet

func (s *PetStoreStdService) AddPet(p Pet, response *AddPetRes) (err error) {
	if s.pets == nil {
		s.pets = []*Pet{}
	}
	newPet := &Pet{}
	*newPet = p
	s.pets = append(s.pets, newPet)
	petRes := AddPetRes(p)
	*response = petRes
	return nil
}

type StandardOpenRPCServiceProvider struct {
	service *PetStoreStdService
}

func (s *StandardOpenRPCServiceProvider) Methods() map[string][]reflect.Value {
	result := make(map[string][]reflect.Value)
	rcvr := reflect.ValueOf(s.service)
	fmt.Println("rcvr val", rcvr)
	for n := 0; n < rcvr.NumMethod(); n++ {
		m := reflect.TypeOf(s.service).Method(n)
		methodName := rcvr.Elem().Type().Name()+"."+m.Name
		fmt.Println("method name", methodName)
		result[methodName] = []reflect.Value{reflect.ValueOf(s.service), m.Func}
	}
	return result
}

func (s *StandardOpenRPCServiceProvider) OpenRPCInfo() types.Info {
	return types.Info{
		Title:          "Test service for pet store",
		Description:    "Which is starting to seem like an actually kind of barbaric thing to have computers involved with...",
		TermsOfService: "Great risk, great reward.",
		Contact:        types.Contact{},
		License:        types.License{},
		Version:        "v0.0.0",
	}
}

func (s *StandardOpenRPCServiceProvider) OpenRPCExternalDocs() types.ExternalDocs {
	return types.ExternalDocs{}
}

type StandardDiscoverArgs string
type StandardDiscoverRes map[string]interface{}

var errorType   = reflect.TypeOf((*error)(nil)).Elem()

func (s *PetStoreStdService) Discover(args StandardDiscoverArgs, response *StandardDiscoverRes) error {
	// Pick from available options.
	opts := &openrpc_go_document.DocumentDiscoverOpts{
		Inline:          false,
		SchemaMutations: []openrpc_go_document.MutateType{openrpc_go_document.SchemaMutateType_Expand, openrpc_go_document.SchemaMutateType_RemoveDefinitions},
		TypeMapper: func(r reflect.Type) *jsonschema.Type {
			switch r {
			case errorType:
				return &jsonschema.Type{
					Type: "error",
				}
			}
			return nil
		},
		IgnoredTypes: []interface{new(error)},
	}

	metaService := &StandardOpenRPCServiceProvider{service:s}
	doc := openrpc_go_document.Wrap(metaService, opts)
	if doc == nil {
		log.Fatal("doc is nil")
	}
	out, err := doc.Discover()
	if err != nil {
		return err
	}
	b, err := json.Marshal(out)
	if err != nil {
		return err
	}
	fmt.Println("doc json raw", string(b))
	err = json.Unmarshal(b, response)
	if err != nil {
		return err
	}
	return nil
}
