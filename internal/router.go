package internal

import (
	"log"
	"strings"
)

// Router dispatches Slack events to matching handlers.
type Router struct {
	handlers       []Handler
	defaultHandler Handler
	prefix         string
}

// NewRouter creates a command router with the given prefix.
func NewRouter(prefix string) *Router {
	if prefix == "" {
		prefix = "!"
	}
	return &Router{prefix: prefix}
}

// Register adds a handler to the router.
func (r *Router) Register(h Handler) {
	r.handlers = append(r.handlers, h)
}

// SetDefault sets the fallback handler for unmatched prefix commands.
func (r *Router) SetDefault(h Handler) {
	r.defaultHandler = h
}

// RouteMessage dispatches a message event to the matching prefix handler.
func (r *Router) RouteMessage(msg MessageEvent, api HandlerAPI) {
	if msg.BotID != "" {
		return
	}
	if !api.Config().IsAllowed(msg.User) {
		return
	}
	content := strings.TrimSpace(msg.Text)
	if !strings.HasPrefix(content, r.prefix) {
		return
	}
	content = strings.TrimPrefix(content, r.prefix)

	for _, h := range r.handlers {
		if h.MatchesPrefix(content) {
			log.Printf("[slack] routing prefix command to %s", h.Name())
			go h.HandleMessage(&msg, api)
			return
		}
	}

	if r.defaultHandler != nil && content != "" {
		msg.Text = "chat " + content
		go r.defaultHandler.HandleMessage(&msg, api)
	}
}

// RouteDirect dispatches a message directly to the default (chat) handler,
// bypassing the prefix requirement. Used for @mentions and DMs.
func (r *Router) RouteDirect(msg MessageEvent, api HandlerAPI) {
	if msg.BotID != "" {
		return
	}
	if !api.Config().IsAllowed(msg.User) {
		return
	}
	content := strings.TrimSpace(msg.Text)
	// Strip bot mention tags like <@U12345> from the beginning
	if idx := strings.Index(content, "> "); idx != -1 && strings.HasPrefix(content, "<@") {
		content = strings.TrimSpace(content[idx+2:])
	} else if strings.HasPrefix(content, "<@") && strings.HasSuffix(content, ">") {
		// Message is just a mention with no text
		content = ""
	}

	if content == "" {
		return
	}

	if r.defaultHandler != nil {
		// Prefix with "chat " so ChatHandler.HandleMessage parses the prompt
		msg.Text = "chat " + content
		log.Printf("[slack] routing direct message to %s", r.defaultHandler.Name())
		go r.defaultHandler.HandleMessage(&msg, api)
	}
}

// RouteSlashCommand dispatches a slash command to the matching handler.
func (r *Router) RouteSlashCommand(cmd SlashCommandPayload, api HandlerAPI) {
	if !api.Config().IsAllowed(cmd.UserID) {
		return
	}
	for _, h := range r.handlers {
		if h.MatchesSlash(cmd.Command) {
			log.Printf("[slack] routing slash command %s to %s", cmd.Command, h.Name())
			go h.HandleSlashCommand(&cmd, api)
			return
		}
	}
}

// RouteInteraction dispatches a block_actions interaction to the matching handler.
func (r *Router) RouteInteraction(payload InteractionPayload, api HandlerAPI) {
	if !api.Config().IsAllowed(payload.User.ID) {
		return
	}
	for _, action := range payload.Actions {
		for _, h := range r.handlers {
			if ih, ok := h.(InteractionHandler); ok && ih.MatchesActionID(action.ActionID) {
				go ih.HandleInteraction(&payload, api)
				return
			}
		}
	}
}
