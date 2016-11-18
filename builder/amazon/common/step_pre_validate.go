package common

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// StepPreValidate provides an opportunity to pre-validate any configuration for
// the build before actually doing any time consuming work
//
type StepPreValidate struct {
	DestAmiName     string
	Owners          []string
	ForceDeregister bool
}

func (s *StepPreValidate) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	if s.ForceDeregister {
		ui.Say("Force Deregister flag found, skipping prevalidating AMI Name")
		return multistep.ActionContinue
	}

	ec2conn := state.Get("ec2").(*ec2.EC2)

	params := &ec2.DescribeImagesInput{}

	params.Filters = []*ec2.Filter{{
		Name:   aws.String("name"),
		Values: []*string{aws.String(s.DestAmiName)},
	}}

	// We have owners to apply
	if len(s.Owners) > 0 {
		owners := make([]*string, len(s.Owners))
		for i, owner := range s.Owners {
			owners[i] = aws.String(owner)
		}
		params.Owners = owners
	}

	ui.Say("Prevalidating AMI Name...")
	resp, err := ec2conn.DescribeImages(params)

	if err != nil {
		err := fmt.Errorf("Error querying AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(resp.Images) > 0 {
		err := fmt.Errorf("Error: name conflicts with an existing AMI: %s", *resp.Images[0].ImageId)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepPreValidate) Cleanup(multistep.StateBag) {}
