package world

import (
	"context"
)

type ConfigurationStruct struct {
	applier  Applier
	childs   []Configuration
	validate func(ctx context.Context) error
}

func NewConfiguration() *ConfigurationStruct {
	return new(ConfigurationStruct)
}

func (c *ConfigurationStruct) WithChilds(childs []Configuration) *ConfigurationStruct {
	c.childs = childs
	return c
}

func (c *ConfigurationStruct) WithApplier(applier Applier) *ConfigurationStruct {
	c.applier = applier
	return c
}

func (c *ConfigurationStruct) WithValidate(validate func(ctx context.Context) error) *ConfigurationStruct {
	c.validate = validate
	return c
}

func (c *ConfigurationStruct) Childs() []Configuration {
	return c.childs
}

func (c *ConfigurationStruct) Applier() Applier {
	return c.applier
}

func (c *ConfigurationStruct) Validate(ctx context.Context) error {
	if c.validate != nil {
		return c.validate(ctx)
	}
	return nil
}
