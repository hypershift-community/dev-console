package ast

import (
	"gopkg.in/yaml.v3"

	"github.com/hypershift-community/hyper-console/pkg/task/errors"
)

type Prompt []string

func (p *Prompt) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		var str string
		if err := node.Decode(&str); err != nil {
			return errors.NewTaskfileDecodeError(err, node)
		}
		*p = []string{str}
		return nil
	case yaml.SequenceNode:
		var list []string
		if err := node.Decode(&list); err != nil {
			return errors.NewTaskfileDecodeError(err, node)
		}
		*p = list
		return nil
	}
	return errors.NewTaskfileDecodeError(nil, node).WithTypeMessage("prompt")
}
