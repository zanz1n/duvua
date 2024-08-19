package manager

import (
	"fmt"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua-bot/internal/errors"
)

type InteractionCreate struct {
	State InteractionState
	*discordgo.Interaction
}

type InteractionState struct {
	Replied bool
	sync.Mutex
}

func newInteractionCreate(i *discordgo.Interaction) *InteractionCreate {
	return &InteractionCreate{
		State:       newInteractionState(),
		Interaction: i,
	}
}

func newInteractionState() InteractionState {
	return InteractionState{
		Replied: false,
		Mutex:   sync.Mutex{},
	}
}

type InteractionResponse struct {
	Content         string
	Components      []discordgo.MessageComponent
	Embeds          []*discordgo.MessageEmbed
	Files           []*discordgo.File
	Attachments     []*discordgo.MessageAttachment
	AllowedMentions *discordgo.MessageAllowedMentions
}

func (r *InteractionResponse) toInterationResponse() *discordgo.InteractionResponse {
	var attachments *[]*discordgo.MessageAttachment = nil
	if r.Attachments != nil {
		attachments = &r.Attachments
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:         r.Content,
			Components:      r.Components,
			Embeds:          r.Embeds,
			Files:           r.Files,
			Attachments:     attachments,
			AllowedMentions: r.AllowedMentions,
		},
	}
}

func (r *InteractionResponse) toWebhookEdit() *discordgo.WebhookEdit {
	var content *string = nil
	if r.Content != "" {
		content = &r.Content
	}
	var embeds *[]*discordgo.MessageEmbed = nil
	if r.Embeds != nil {
		embeds = &r.Embeds
	}
	var components *[]discordgo.MessageComponent = nil
	if r.Components != nil {
		components = &r.Components
	}
	var attachments *[]*discordgo.MessageAttachment = nil
	if r.Attachments != nil {
		attachments = &r.Attachments
	}

	return &discordgo.WebhookEdit{
		Content:         content,
		Components:      components,
		Embeds:          embeds,
		Files:           r.Files,
		Attachments:     attachments,
		AllowedMentions: r.AllowedMentions,
	}
}

func (i *InteractionCreate) Replied() bool {
	i.State.Lock()
	replied := i.State.Replied
	i.State.Unlock()

	return replied
}

func (i *InteractionCreate) Replyf(s *discordgo.Session, f string, a ...any) error {
	return i.Reply(s, &InteractionResponse{Content: fmt.Sprintf(f, a...)})
}

func (i *InteractionCreate) ReplyEphemeralf(s *discordgo.Session, f string, a ...any) error {
	return i.ReplyEphemeral(s, &InteractionResponse{Content: fmt.Sprintf(f, a...)})
}

func (i *InteractionCreate) DeferReply(s *discordgo.Session, ephemeral bool) error {
	i.State.Lock()
	defer i.State.Unlock()

	flags := discordgo.MessageFlags(0)
	if ephemeral {
		flags |= discordgo.MessageFlagsEphemeral
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: flags},
	})
	if err == nil {
		i.State.Replied = true
	}

	return err
}

func (i *InteractionCreate) DeferUpdate(s *discordgo.Session) error {
	i.State.Lock()
	defer i.State.Unlock()

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err == nil {
		i.State.Replied = true
	}

	return err
}

func (i *InteractionCreate) Reply(s *discordgo.Session, resp *InteractionResponse) error {
	i.State.Lock()
	defer i.State.Unlock()

	if i.State.Replied {
		_, err := s.InteractionResponseEdit(i.Interaction, resp.toWebhookEdit())
		return err
	} else {
		return s.InteractionRespond(i.Interaction, resp.toInterationResponse())
	}
}

func (i *InteractionCreate) ReplyEphemeral(s *discordgo.Session, resp *InteractionResponse) error {
	i.State.Lock()
	defer i.State.Unlock()

	if i.State.Replied {
		_, err := s.InteractionResponseEdit(i.Interaction, resp.toWebhookEdit())
		return err
	} else {
		res := resp.toInterationResponse()
		res.Data.Flags = discordgo.MessageFlagsEphemeral
		return s.InteractionRespond(i.Interaction, res)
	}
}

func (i *InteractionCreate) GetTypedOption(
	name string,
	required bool,
	kind discordgo.ApplicationCommandOptionType,
) (*discordgo.ApplicationCommandInteractionDataOption, error) {
	opt, err := i.GetOption(name, required)
	if err != nil {
		return nil, err
	} else if opt == nil {
		return nil, nil
	}

	if opt.Type != kind {
		return nil, errors.Newf("opção `%s` de tipo inesperado", opt.Name)
	}

	return opt, nil
}

func (i *InteractionCreate) GetOption(
	name string,
	required bool,
) (*discordgo.ApplicationCommandInteractionDataOption, error) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return nil, errors.New("interação de tipo inesperado")
	}
	data := i.ApplicationCommandData()

	var opts []*discordgo.ApplicationCommandInteractionDataOption
	if len(data.Options) == 1 {
		switch data.Options[0].Type {
		case discordgo.ApplicationCommandOptionSubCommand:
			opts = data.Options[0].Options

		case discordgo.ApplicationCommandOptionSubCommandGroup:
			if len(data.Options[0].Options) != 1 {
				return nil, errors.Newf("opção `%s` é necessária", name)
			}
			opts = data.Options[0].Options[0].Options
		}
	} else {
		opts = data.Options
	}

	for _, opt := range opts {
		if opt.Name == name {
			return opt, nil
		}
	}

	if required {
		return nil, errors.Newf("opção `%s` é necessária", name)
	}
	return nil, nil
}

func (i *InteractionCreate) GetSubCommand() (
	*discordgo.ApplicationCommandInteractionDataOption,
	error,
) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return nil, errors.New("interação de tipo inesperado")
	}
	data := i.ApplicationCommandData()

	if len(data.Options) != 1 {
		return nil, errors.New("opção `sub-command` é necessária")
	}

	switch data.Options[0].Type {
	case discordgo.ApplicationCommandOptionSubCommand:
		return data.Options[0], nil

	case discordgo.ApplicationCommandOptionSubCommandGroup:
		if len(data.Options[0].Options) != 1 {
			return nil, errors.New("opção `sub-command` é necessária")
		}
		t := data.Options[0].Options[0].Type
		if t != discordgo.ApplicationCommandOptionSubCommand {
			return nil, errors.New("opção `sub-command` é necessária")
		}
		return data.Options[0].Options[0], nil
	default:
		return nil, errors.New("opção `sub-command` é necessária")
	}
}

func (i *InteractionCreate) GetSubCommandGroup() *discordgo.ApplicationCommandInteractionDataOption {
	if i.Type != discordgo.InteractionApplicationCommand {
		return nil
	}
	data := i.ApplicationCommandData()

	if len(data.Options) != 1 {
		return nil
	}

	t := data.Options[0].Type
	if t != discordgo.ApplicationCommandOptionSubCommandGroup {
		return nil
	}

	return data.Options[0]
}

func (i *InteractionCreate) GetStringOption(name string, required bool) (string, error) {
	opt, err := i.GetTypedOption(name, required, discordgo.ApplicationCommandOptionString)
	if err != nil {
		return "", err
	} else if opt == nil {
		return "", nil
	}

	return opt.StringValue(), nil
}

func (i *InteractionCreate) GetIntegerOption(
	name string,
	required bool,
	defaultValue ...int64,
) (int64, error) {
	opt, err := i.GetTypedOption(name, required, discordgo.ApplicationCommandOptionInteger)
	if err != nil {
		return 0, err
	} else if opt == nil {
		if len(defaultValue) > 1 {
			return defaultValue[0], nil
		}
		return 0, nil
	}

	return opt.IntValue(), nil
}

func (i *InteractionCreate) GetBooleanOption(
	name string,
	required bool,
	defaultValue ...bool,
) (bool, error) {
	opt, err := i.GetTypedOption(name, required, discordgo.ApplicationCommandOptionBoolean)
	if err != nil {
		return false, err
	} else if opt == nil {
		if len(defaultValue) > 1 {
			return defaultValue[0], nil
		}
		return false, nil
	}

	return opt.BoolValue(), nil
}

func (i *InteractionCreate) GetUserOption(name string, required bool) (string, error) {
	opt, err := i.GetTypedOption(name, required, discordgo.ApplicationCommandOptionUser)
	if err != nil {
		return "", err
	} else if opt == nil {
		return "", nil
	}
	userId := opt.Value.(string)

	return userId, nil
}

func (i *InteractionCreate) GetChannelOption(name string, required bool) (string, error) {
	opt, err := i.GetTypedOption(name, required, discordgo.ApplicationCommandOptionChannel)
	if err != nil {
		return "", err
	} else if opt == nil {
		return "", nil
	}
	channelId := opt.Value.(string)

	return channelId, nil
}
