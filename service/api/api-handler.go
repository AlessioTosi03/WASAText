package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// Register routes
	rt.router.GET("/files/:photo", rt.serveFile)
	rt.router.PUT("/users/:UserID/name", rt.setMyUserName)
	rt.router.PUT("/users/:UserID/photo", rt.setMyPhoto)
	rt.router.POST("/chat/:ConversationID", rt.addToGroup)
	rt.router.GET("/context", rt.wrap(rt.getContextReply))
	rt.router.DELETE("/chat/:ConversationID", rt.leaveGroup)
	rt.router.PUT("/chat/:ConversationID/name", rt.setGroupName)
	rt.router.PUT("/chat/:ConversationID/photo", rt.setGroupPhoto)
	rt.router.GET("/chat/:ConversationID/messages", rt.getConversation)
	rt.router.POST("/chat/:ConversationID/messages", rt.sendMessage)
	rt.router.POST("/chat/:ConversationID/forward", rt.forwardMessage)
	rt.router.GET("/chat/:ConversationID/messages/:MessageID/reaction", rt.getMyReaction)
	rt.router.DELETE("/chat/:ConversationID/messages/:MessageID", rt.deleteMessage)
	rt.router.POST("/chat/:ConversationID/messages/:MessageID/reaction", rt.commentMessage)
	rt.router.DELETE("/chat/:ConversationID/messages/:MessageID/reaction", rt.uncommentMessage)
	rt.router.GET("/stream", rt.getMyConversations)
	rt.router.POST("/session", rt.doLogin)
	rt.router.POST("/newGroup", rt.createGroup)
	rt.router.POST("/newChat", rt.createChat)
	rt.router.GET("/users", rt.getAllUsers)
	rt.router.GET("/users/:username", rt.getUserIDbyUsername)
	rt.router.GET("/chat/:ConversationID/participants", rt.getGroupParticipants)
	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}
