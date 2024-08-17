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

func (i *InteractionCreate) GetTypedOption(
	name string,
	required bool,
	kind discordgo.ApplicationCommandOptionType,
) (*discordgo.ApplicationCommandInteractionDataOption, error) {
	opt, err := i.GetOption(name, required)
	if err != nil {
		return nil, err
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

	for _, opt := range data.Options {
		if opt.Type == discordgo.ApplicationCommandOptionSubCommand {
			return opt, nil
		}
	}

	return nil, errors.New("opção `sub-command` é necessária")
}
