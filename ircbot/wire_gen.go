// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package ircbot

import (
	"context"
	"errors"
	"fmt"
	"github.com/R-a-dio/valkyrie/database"
	"github.com/R-a-dio/valkyrie/rpc/manager"
	"github.com/davecgh/go-spew/spew"
	"github.com/google/wire"
	"github.com/jmoiron/sqlx"
	"github.com/lrstanley/girc"
	"strconv"
)

// Injectors from providers.go:

func NowPlaying(event Event) (CommandFn, error) {
	client := WithClient(event)
	gircEvent := WithIRCEvent(event)
	respondPublic := WithRespondPublic(client, gircEvent)
	bot := WithBot(event)
	statusResponse, err := WithManagerStatus(bot)
	if err != nil {
		return nil, err
	}
	db := WithDB(bot)
	handler := WithDatabase(db)
	currentTrack, err := WithCurrentTrack(handler, statusResponse)
	if err != nil {
		return nil, err
	}
	commandFn := nowPlaying(respondPublic, statusResponse, currentTrack)
	return commandFn, nil
}

func TrackTags(event Event) (CommandFn, error) {
	client := WithClient(event)
	gircEvent := WithIRCEvent(event)
	respond := WithRespond(client, gircEvent)
	bot := WithBot(event)
	db := WithDB(bot)
	handler := WithDatabase(db)
	arguments := WithArguments(event)
	statusResponse, err := WithManagerStatus(bot)
	if err != nil {
		return nil, err
	}
	argumentOrCurrentTrack, err := WithArgumentOrCurrentTrack(handler, arguments, statusResponse)
	if err != nil {
		return nil, err
	}
	commandFn := trackTags(respond, argumentOrCurrentTrack)
	return commandFn, nil
}

// providers.go:

var Providers = wire.NewSet(
	WithBot,
	WithDB,
	WithClient,
	WithIRCEvent,

	WithDatabase,
	WithDatabaseTx,
	WithManagerStatus,

	WithArguments,

	WithCurrentTrack,
	WithArgumentTrack,
	WithArgumentOrCurrentTrack,

	WithRespond,
	WithRespondPrivate,
	WithRespondPublic,
)

type CommandFn func(*girc.Client, girc.Event) error

type Arguments map[string]string

func WithBot(e Event) *Bot {
	return e.bot
}

func WithDB(bot *Bot) *sqlx.DB {
	return bot.DB
}

func WithIRCEvent(e Event) girc.Event {
	return e.e
}

func WithClient(e Event) *girc.Client {
	return e.c
}

func WithArguments(e Event) Arguments {
	return e.a
}

func WithManagerStatus(bot *Bot) (*manager.StatusResponse, error) {
	return bot.Manager.Status(context.TODO(), new(manager.StatusRequest))
}

func WithDatabase(db *sqlx.DB) database.Handler {
	return database.Handle(context.TODO(), db)
}

func WithDatabaseTx(db *sqlx.DB) (database.HandlerTx, error) {
	return database.HandleTx(context.TODO(), db)
}

// ArgumentTrack is a track specified by user arguments
type ArgumentTrack struct {
	database.Track
}

// WithArgumentTrack returns the track specified by the argument 'TrackID'
func WithArgumentTrack(h database.Handler, a Arguments) (ArgumentTrack, error) {
	spew.Dump(a)
	id := a["TrackID"]
	if id == "" {
		return ArgumentTrack{}, errors.New("no trackid found in arguments")
	}

	tid, err := strconv.Atoi(id)
	if err != nil {
		panic("non-numeric TrackID found from arguments: " + id)
	}

	track, err := database.GetTrack(h, database.TrackID(tid))
	return ArgumentTrack{track}, err
}

// CurrentTrack is a track that is currently being played on stream
type CurrentTrack struct {
	database.Track
}

// WithCurrentTrack returns the currently playing track
func WithCurrentTrack(h database.Handler, s *manager.StatusResponse) (CurrentTrack, error) {
	track, err := database.GetSongFromMetadata(h, s.Song.Metadata)
	return CurrentTrack{track}, err
}

type ArgumentOrCurrentTrack struct {
	database.Track
}

// WithArgumentOrCurrentTrack combines WithArgumentTrack and WithCurrentTrack returning
// ArgumentTrack first if available
func WithArgumentOrCurrentTrack(h database.Handler, a Arguments, s *manager.StatusResponse) (ArgumentOrCurrentTrack, error) {
	trackA, err := WithArgumentTrack(h, a)
	if err == nil {
		return ArgumentOrCurrentTrack{trackA.Track}, err
	} else if err == database.ErrTrackNotFound {
		return ArgumentOrCurrentTrack{}, NewUserError(err, "track identifier does not exist")
	}

	trackC, err := WithCurrentTrack(h, s)
	return ArgumentOrCurrentTrack{trackC.Track}, err
}

type Respond func(message string, args ...interface{})

type RespondPrivate func(message string, args ...interface{})

type RespondPublic func(message string, args ...interface{})

func WithRespond(c *girc.Client, e girc.Event) Respond {
	switch e.Trailing[0] {
	case '.', '!':
		return Respond(WithRespondPrivate(c, e))
	case '@':
		return Respond(WithRespondPublic(c, e))
	default:
		panic("non-prefixed regular expression used")
	}
}

func WithRespondPrivate(c *girc.Client, e girc.Event) RespondPrivate {
	return func(message string, args ...interface{}) {
		message = girc.Fmt(message)
		message = fmt.Sprintf(message, args...)

		c.Cmd.Notice(e.Source.Name, message)
	}
}

func WithRespondPublic(c *girc.Client, e girc.Event) RespondPublic {
	return func(message string, args ...interface{}) {
		message = girc.Fmt(message)
		message = fmt.Sprintf(message, args...)

		c.Cmd.Message(e.Params[0], message)
	}
}

func LastPlayed(Event) (CommandFn, error) {

	return nil, nil
}

func StreamerQueue(Event) (CommandFn, error) {

	return nil, nil
}

func StreamerQueueLength(Event) (CommandFn, error) {

	return nil, nil
}

func StreamerUserInfo(Event) (CommandFn, error) {

	return nil, nil
}

func FaveTrack(Event) (CommandFn, error) {

	return nil, nil
}

func FaveList(Event) (CommandFn, error) {

	return nil, nil
}

func ThreadURL(Event) (CommandFn, error) {

	return nil, nil
}

func ChannelTopic(Event) (CommandFn, error) {

	return nil, nil
}

func KillStreamer(Event) (CommandFn, error) {

	return nil, nil
}

func RandomTrackRequest(Event) (CommandFn, error) {

	return nil, nil
}

func LuckyTrackRequest(Event) (CommandFn, error) {

	return nil, nil
}

func SearchTrack(Event) (CommandFn, error) {

	return nil, nil
}

func RequestTrack(Event) (CommandFn, error) {

	return nil, nil
}

func LastRequestInfo(Event) (CommandFn, error) {

	return nil, nil
}

func TrackInfo(Event) (CommandFn, error) {

	return nil, nil
}
