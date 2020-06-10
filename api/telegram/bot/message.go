// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

import "strings"

//
// Message represents a message.
//
type Message struct {
	MessageForward

	ID   int   `json:"message_id"` // Unique message identifier inside this chat.
	Date int   `json:"date"`       // Date the message was sent in Unix time.
	Chat *Chat `json:"chat"`       // Conversation the message belongs to.

	// Optional. Sender, empty for messages sent to channels.
	From *User `json:"from"`

	// Optional. Date the message was last edited in Unix time
	EditDate int `json:"edit_date"`

	// Optional. For replies, the original message.
	// Note that the Message object in this field will not contain further
	// reply_to_message fields even if it itself is a reply.
	ReplyTo *Message `json:"reply_to_message"`

	// Optional. The unique identifier of a media message group this
	// message belongs to.
	MediaGroupID string `json:"media_group_id"`

	// Optional. Signature of the post author for messages in channels.
	AuthorSignature string `json:"author_signature"`

	// Optional. For text messages, the actual UTF-8 text of the message,
	// 0-4096 characters.
	Text string `json:"text"`

	// Optional. For text messages, special entities like usernames, URLs,
	// bot commands, etc. that appear in the text.
	Entities []MessageEntity `json:"entities"`

	// Optional. Message is an audio file, information about the file
	Audio *Audio `json:"audio"`

	// Optional. Message is a general file, information about the file.
	Document *Document `json:"document"`

	// Optional. Message is an animation, information about the animation.
	// For backward compatibility, when this field is set, the document
	// field will also be set.
	Animation *Animation `json:"animation"`

	// Optional. Message is a game, information about the game.
	Game *Game `json:"game"`

	// Optional. Message is a photo, available sizes of the photo.
	Photo []PhotoSize `json:"photo"`

	// Optional. Message is a sticker, information about the sticker.
	Sticker *Sticker `json:"sticker"`

	// Optional. Message is a video, information about the video.
	Video *Video `json:"video"`

	// Optional. Message is a voice message, information about the file.
	Voice *Voice `json:"voice"`

	// Optional. Message is a video note, information about the video
	// message.
	VideoNote *VideoNote `json:"video_note"`

	// Optional. Caption for the animation, audio, document, photo, video
	// or voice, 0-1024 characters.
	Caption string `json:"caption"`

	// Optional. For messages with a caption, special entities like
	// usernames, URLs, bot commands, etc. that appear in the caption.
	CaptionEntities []MessageEntity `json:"caption_entities"`

	// Optional. Message is a shared contact, information about the
	// contact.
	Contact *Contact `json:"contact"`

	// Optional. Message is a shared location, information about the
	// location.
	Location *Location `json:"location"`

	// Optional. Message is a venue, information about the venue.
	Venue *Venue `json:"venue"`

	// Optional. Message is a native poll, information about the poll.
	Poll *Poll `json:"poll"`

	// Optional. Message is a dice with random value from 1 to 6.
	Dice *Dice `json:"dice"`

	// Optional. New members that were added to the group or supergroup
	// and information about them (the bot itself may be one of these
	// members).
	NewMembers []*User `json:"new_chat_members"`

	// Optional. A member was removed from the group, information about
	// them (this member may be the bot itself).
	LeftMembers []*User `json:"left_chat_members"`

	// Optional. A chat title was changed to this value.
	NewChatTitle string `json:"new_chat_title"`

	// Optional. A chat photo was change to this value.
	NewChatPhoto []PhotoSize `json:"new_chat_photo"`

	// Optional. Service message: the chat photo was deleted.
	IsChatPhotoDeleted bool `json:"delete_chat_photo"`

	// Optional. Service message: the group has been created.
	IsGroupChatCreated bool `json:"group_chat_created"`

	// Optional. Service message: the supergroup has been created. This
	// field can‘t be received in a message coming through updates,
	// because bot can’t be a member of a supergroup when it is created.
	// It can only be found in reply_to_message if someone replies to a
	// very first message in a directly created supergroup.
	IsSupergroupChatCreated bool `json:"supergroup_chat_created"`

	// Optional. Service message: the channel has been created.
	// This field can‘t be received in a message coming through updates,
	// because bot can’t be a member of a channel when it is created.
	// It can only be found in reply_to_message if someone replies to a
	// very first message in a channel.
	IsChannelChatCreated bool `json:"channel_chat_created"`

	// Optional. The group has been migrated to a supergroup with the
	// specified identifier.
	MigrateToChatID int64 `json:"migrate_to_chat_id"`

	// Optional. The supergroup has been migrated from a group with the
	// specified identifier.
	MigrateFromChatID int64 `json:"migrate_from_chat_id"`

	// Optional. Specified message was pinned.
	// Note that the Message object in this field will not contain further
	// reply_to_message fields even if it is itself a reply.
	PinnedMessage *Message `json:"pinned_message"`

	// Optional. Message is an invoice for a payment, information about
	// the invoice.
	Invoice *Invoice `json:"invoice"`

	// Optional. Message is a service message about a successful payment,
	// information about the payment.
	SuccessfulPayment *SuccessfulPayment `json:"successful_payment"`

	// Optional. The domain name of the website on which the user has
	// logged in.
	ConnectedWebsite string `json:"connected_website"`

	// Optional. Telegram Passport data.
	PassportData *PassportData `json:"passport_data"`

	// Optional. Inline keyboard attached to the message.
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup"`

	Command     string // It will contains the Command name.
	CommandArgs string // It will contains the Command arguments.
}

//
// parseCommandArgs parse the Text to get the command and its arguments.
//
func (msg *Message) parseCommandArgs() bool {
	var cmdEntity *MessageEntity

	for x, ent := range msg.Entities {
		if ent.Type == EntityTypeBotCommand {
			cmdEntity = &msg.Entities[x]
			break
		}
	}
	if cmdEntity == nil {
		return false
	}

	start := cmdEntity.Offset
	end := start + cmdEntity.Length

	msg.Command = strings.TrimPrefix(msg.Text[start:end], "/")
	msg.Command = strings.Split(msg.Command, "@")[0]
	msg.CommandArgs = strings.TrimSpace(msg.Text[end:])

	return true
}
