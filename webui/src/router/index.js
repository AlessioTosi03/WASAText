import {createRouter, createWebHashHistory} from 'vue-router'
import HomeView from '../views/HomeView.vue'
import NewChatView from '../views/NewChatView.vue'
import NewGroupView from '../views/NewGroupView.vue'

const router = createRouter({
	history: createWebHashHistory(import.meta.env.BASE_URL),
	routes: [
		{path: '/', component: HomeView},
		{path: '/newChat', component: NewChatView},
		{path: '/newGroup', component: NewGroupView},
		{path: '/session', component: HomeView},
		{path: '/some/:id/link', component: HomeView},
	]
})

export default router
