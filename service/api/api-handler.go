package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// Register routes
	rt.router.PUT("/Users/:UserID/name", rt.setMyUserName)
	rt.router.PUT("/Users/:UserID/photo", rt.setMyPhoto)
	rt.router.POST("/chat/:ConversationID", rt.addToGroup)
	rt.router.GET("/context", rt.wrap(rt.getContextReply))
	rt.router.DELETE("/chat/:ConversationID", rt.leaveGroup)
	rt.router.PUT("/chat/:ConversationID/name", rt.setGroupName)
	rt.router.PUT("/chat/:ConversationID/photo", rt.setGroupPhoto)
	rt.router.GET("/chat/:ConversationID/messages", rt.getConversation)
	rt.router.POST("/chat/:ConversationID/messages", rt.sendMessages)
	rt.router.POST("/chat/:ConversationID/forward", rt.forwardMessages)
	rt.router.DELETE("/chat/:ConversationID/messages/:MessageID", rt.deleteMessage)
	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}
