package commands

import (
	"eoracle-client-server/internal/output"
	"eoracle-client-server/internal/storage"
	"errors"
	"fmt"
	"log"
	"strings"
)

const (
	AddItem     CommandType = "addItem"
	DeleteItem  CommandType = "deleteItem"
	GetItem     CommandType = "getItem"
	GetAllItems CommandType = "getAllItems"
)

const (
	ReadCategory  CommandCategory = "READ"
	WriteCategory CommandCategory = "WRITE"
)

type CommandRegistry interface {
	ParseCommand(line string) (Command, error)
	HandleCommand(cmd Command, store storage.Storage, out output.Output) error
	IsReadCommand(cmd Command) bool
}

type commandRegistry struct {
	byName map[string]CommandSpec
	byType map[CommandType]CommandSpec
}

type CommandSpec struct {
	Type     CommandType
	Category CommandCategory
	Parser   func(args []string) (Command, error)
	Handler  func(cmd Command, store storage.Storage, out output.Output) error
}

func NewCommandRegistry() CommandRegistry {
	commandRegistry := &commandRegistry{
		byName: make(map[string]CommandSpec),
		byType: make(map[CommandType]CommandSpec),
	}

	commandRegistry.Register("add", CommandSpec{
		Type:     AddItem,
		Category: WriteCategory,
		Parser: func(args []string) (Command, error) {
			if len(args) < 2 {
				return nil, errors.New("add command requires key and value")
			}
			return &command{
				Type:  AddItem,
				Key:   args[0],
				Value: strings.Join(args[1:], " "),
			}, nil
		},
		Handler: func(cmd Command, store storage.Storage, out output.Output) error {
			store.Add(cmd.GetKey(), cmd.GetValue())
			log.Printf("Added item: %s = %s", cmd.GetKey(), cmd.GetValue())
			return nil
		},
	})

	commandRegistry.Register("delete", CommandSpec{
		Type:     DeleteItem,
		Category: WriteCategory,
		Parser: func(args []string) (Command, error) {
			if len(args) < 1 {
				return nil, errors.New("delete command requires key")
			}
			return &command{
				Type: DeleteItem,
				Key:  args[0],
			}, nil
		},
		Handler: func(cmd Command, store storage.Storage, out output.Output) error {
			deleted := store.Delete(cmd.GetKey())
			if deleted {
				log.Printf("Deleted item: %s", cmd.GetKey())
			} else {
				log.Printf("Item not found for deletion: %s", cmd.GetKey())
			}
			return nil
		},
	})

	commandRegistry.Register("get", CommandSpec{
		Type:     GetItem,
		Category: ReadCategory,
		Parser: func(args []string) (Command, error) {
			if len(args) < 1 {
				return nil, errors.New("get command requires key")
			}
			return &command{
				Type: GetItem,
				Key:  args[0],
			}, nil
		},
		Handler: func(cmd Command, store storage.Storage, out output.Output) error {
			value, exists := store.Get(cmd.GetKey())
			if exists {
				out.Write(fmt.Sprintf("%s = %s\n", cmd.GetKey(), value))
				log.Printf("Retrieved item: %s = %s", cmd.GetKey(), value)
			} else {
				log.Printf("Item not found: %s", cmd.GetKey())
			}

			return nil
		},
	})

	commandRegistry.Register("getall", CommandSpec{
		Type:     GetAllItems,
		Category: ReadCategory,
		Parser: func(args []string) (Command, error) {
			return &command{Type: GetAllItems}, nil
		},
		Handler: func(cmd Command, store storage.Storage, out output.Output) error {
			allItems := store.GetAll()
			writeData := ""
			for _, item := range allItems {
				writeData += fmt.Sprintf("%s = %s\n", item.Key, item.Value)
			}
			out.Write(writeData)
			log.Printf("Retrieved all items: %d items", len(allItems))
			return nil
		},
	})

	return commandRegistry
}

func (r commandRegistry) Register(name string, spec CommandSpec) {
	r.byName[strings.ToLower(name)] = spec
	r.byType[spec.Type] = spec
}

func (r commandRegistry) ParseCommand(line string) (Command, error) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil, errors.New("empty command")
	}
	name := strings.ToLower(parts[0])
	args := parts[1:]

	spec, ok := r.byName[name]
	if !ok {
		return nil, fmt.Errorf("unknown command: %s", name)
	}

	return spec.Parser(args)
}

func (r commandRegistry) HandleCommand(cmd Command, store storage.Storage, out output.Output) error {
	spec, ok := r.byType[cmd.GetType()]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.GetType())
	}

	return spec.Handler(cmd, store, out)
}

func (r commandRegistry) IsReadCommand(cmd Command) bool {
	spec, ok := r.byType[cmd.GetType()]
	if !ok {
		return false
	}

	return spec.Category == ReadCategory
}
