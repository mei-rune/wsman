package winrm

import (
	"errors"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/runner-mei/wsman"
	"github.com/runner-mei/wsman/envelope"
	"io"
	"launchpad.net/xmlpath"
	"log"
)

type Shell struct {
	Id       string
	Endpoint string
	Owner    string
	password string
	Stdout   io.Writer
	Stderr   io.Writer
}

func NewShell(endpoint, user, pass string) (*Shell, error) {
	env := &envelope.CreateShell{uuid.TimeOrderedUUID()}
	response, err := wsman.Deliver(endpoint, user, pass, env)
	if err != nil {
		return nil, err
	}

	path := xmlpath.MustCompile("//Body/ResourceCreated/ReferenceParameters/SelectorSet/Selector[@Name='ShellId']")
	root, err := xmlpath.Parse(response)
	if err != nil {
		return nil, err
	}

	id, ok := path.String(root)
	if !ok {
		return nil, errors.New("Could not create shell.")
	}

	return &Shell{id, endpoint, user, pass, nil, nil}, nil
}

func (s *Shell) NewCommand(cmd string) (*Command, error) {
	env := &envelope.CreateCommand{uuid.TimeOrderedUUID(), s.Id, cmd}
	response, err := wsman.Deliver(s.Endpoint, s.Owner, s.password, env)
	if err != nil {
		return nil, err
	}

	path := xmlpath.MustCompile("//Body/CommandResponse/CommandId")
	root, err := xmlpath.Parse(response)
	if err != nil {
		return nil, err
	}

	id, ok := path.String(root)
	if !ok {
		return nil, errors.New("Could not create command.")
	}

	return &Command{s, id, cmd}, nil
}

func (s *Shell) Delete() error {
	env := &envelope.DeleteShell{uuid.TimeOrderedUUID(), s.Id}
	_, err := wsman.Deliver(s.Endpoint, s.Owner, s.password, env)

	if err != nil {
		log.Println(err.Error())
	}
	return err
}
