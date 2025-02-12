package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// Register routes
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
	rt.router.GET("/chat/:ConversationID/messages/:MessageID/reactions", rt.getMyReaction)
	rt.router.DELETE("/chat/:ConversationID/messages/:MessageID", rt.deleteMessage)
	rt.router.POST("/chat/:ConversationID/messages/:MessageID/reactions", rt.commentMessage)
	rt.router.DELETE("/chat/:ConversationID/messages/:MessageID/reactions", rt.uncommentMessage)
	rt.router.GET("/stream", rt.getMyConversations)
	rt.router.POST("/session", rt.doLogin)
	rt.router.POST("/newGroup", rt.createGroup)
	rt.router.POST("/newChat", rt.createChat)
	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}
