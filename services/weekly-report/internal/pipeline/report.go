package pipeline

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/therealagt/ContractManagementTool/libs/common/email"
	"github.com/therealagt/ContractManagementTool/libs/common/report"
)

type Mailer interface {
	Enabled() bool
	Send(to []string, subject, bodyHTML string) error
}

type Pipeline struct {
	generator *report.Generator
	mailer    Mailer
	env       string
	recipients []string
}

func New(db *sql.DB, mailer Mailer, reviewSLADays int, environment string, recipients []string) *Pipeline {
	return &Pipeline{
		generator:  report.NewGenerator(db, reviewSLADays),
		mailer:     mailer,
		env:        environment,
		recipients: recipients,
	}
}

func (p *Pipeline) Run(ctx context.Context) error {
	status, err := p.generator.WeeklyStatus(ctx)
	if err != nil {
		return fmt.Errorf("weekly status: %w", err)
	}

	if !p.mailer.Enabled() {
		log.Printf("weekly report generated (email disabled): traffic_light=%s uploads=%d archived=%d",
			status.TrafficLight, status.Week.Uploaded, status.Week.Archived)
		return nil
	}

	subject := status.Subject(p.env)
	body := status.HTML()
	if err := p.mailer.Send(p.recipients, subject, body); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	log.Printf("weekly report sent to %d recipients (status=%s)", len(p.recipients), status.TrafficLight)
	return nil
}

func NewMailerFromConfig(host string, port int, user, password, from string) Mailer {
	return email.NewSender(email.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		From:     from,
	})
}
