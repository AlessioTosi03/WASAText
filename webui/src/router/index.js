import {createRouter, createWebHashHistory} from 'vue-router'
import HomeView from '../views/HomeView.vue'
import NewChatView from '../views/NewChatView.vue'
import NewGroupView from '../views/NewGroupView.vue'
import MessagesView from '../views/MessagesView.vue'
import UserView from '../views/UserView.vue'

const router = createRouter({
	history: createWebHashHistory(import.meta.env.BASE_URL),
	routes: [
		{path: '/', component: HomeView},
		{path: '/newChat', component: NewChatView},
		{path: '/newGroup', component: NewGroupView},
		{path: '/session', component: HomeView},
		{path: '/chat/:conversation_id/messages', component: MessagesView},
		{path: '/Users/:user_id', component: UserView}
	]
})

export default router
