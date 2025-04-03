package pinpoint

import (
	"context"
	"dashfun_gamecenter/config"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/pinpointemail"
	"github.com/aws/aws-sdk-go-v2/service/pinpointemail/types"
	"log"
	"sync"
)

var once sync.Once
var inst *Pinpoint

type Pinpoint struct {
	client *pinpointemail.Client
}

func Get() *Pinpoint {
	once.Do(func() {
		p := &Pinpoint{}
		p.init()
		inst = p
	})
	return inst
}

func (p *Pinpoint) init() {
	pc := config.GetConfig().PinPoint
	creds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(pc.KeyId, pc.Secret, ""))
	cfg, err := awscfg.LoadDefaultConfig(context.TODO(), awscfg.WithCredentialsProvider(creds), awscfg.WithRegion("ca-central-1"))
	if err != nil {
		log.Fatalln(err)
	}

	client := pinpointemail.NewFromConfig(cfg)
	p.client = client
}

func (p *Pinpoint) SendEmail(subject string, to string, content string) error {
	input := &pinpointemail.SendEmailInput{
		FromEmailAddress: aws.String("Service<service@metavirus.games>"),
		Destination: &types.Destination{
			ToAddresses: []string{
				to,
			},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(subject),
				},
				Body: &types.Body{
					Text: &types.Content{
						Data:    aws.String(content),
						Charset: aws.String("UTF-8"),
					},
				},
			},
		},
	}

	result, err := p.client.SendEmail(context.Background(), input)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Email sent successfully! Message ID: %s\n", *result.MessageId)
	return err
}
