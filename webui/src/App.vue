<script setup>
import { RouterLink, RouterView } from 'vue-router'
import HomeView from './views//HomeView.vue';
</script>
<script>
export default {
	data: function() {
		return {
			errormsg: null,
			loading: false,
			conversations: [],
		}
	},
	components: {
		HomeView
	},
	methods: {
		async refresh() {
			this.loading = true;
			this.errormsg = null;

			const token = localStorage.getItem("token");  // Get the user_id from storage

			if (!token) {
				this.isLoggedIn = false;  // User is NOT logged in
				this.conversations = [];
				this.errormsg = "No token found. Please log in.";
				this.loading = false;
				return;
			}

			try {
				let response = await this.$axios.get("/stream", {
					headers: { Authorization: `Bearer ${token}` }
				});
				
				// If the request is successful, update conversations
				this.conversations = response.data;
				this.isLoggedIn = true;  // User is logged in
			} catch (e) {
				if (e.response && e.response.status === 401) {
					// Invalid token, user should log in again
					localStorage.removeItem("token");  // Remove bad token
					this.isLoggedIn = false;
					this.errormsg = "Invalid token. Please log in again.";
					this.conversations = [];
				} else {
					// Handle other errors (network issues, etc.)
					this.errormsg = `Error: ${e.response ? e.response.data : e.toString()}`;
				}
			}

			this.loading = false;
		},
	},
	mounted() {
		this.refresh()
	}
}
</script>

<template>

	<header class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow">
		<a class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6" href="#/">Example App</a>
		<button class="navbar-toggler position-absolute d-md-none collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#sidebarMenu" aria-controls="sidebarMenu" aria-expanded="false" aria-label="Toggle navigation">
			<span class="navbar-toggler-icon"></span>
		</button>
	</header>

	<div class="container-fluid">
		<div class="row">
			<nav id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse">
				<div class="position-sticky pt-3 sidebar-sticky">
					<h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted text-uppercase">
						<span>Conversations</span>
					</h6>
					<ul class = "conversations-list">
						<li v-for="(c, index) in conversations" :key="index" class="nav-item">
							<router-link  :to="'/chats/' + c.conversation.id + '/messages'">
									<!-- If it's a group conversation, show the group name -->
									<div v-if="c.conversation.type === 'group'" class = "conversation-item">
										<img :src="c.group.photo" width="40" height="40" class="conversation" id= "conversation-photo">
										<p class="conversation" id="conversation-name">{{ c.group.name }}</p> <!-- Group name -->
									</div>

									<!-- If it's a chat conversation, show the other participant's name -->
									<div v-else class = "conversation-item">
										<img :src="c.other_user.profile_pic" width="40" height="40" class="conversation" id="conversation-photo">
										<p class="conversation" id="conversation-name">{{ c.other_user.username }}</p>
									</div>
							</router-link>
						</li>
					</ul>

					<ul class="nav flex-column">
						<li class="nav-item">
							<RouterLink to="/" class="nav-link">
								<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#home"/></svg>
								Home
							</RouterLink>
						</li>
						<li class="nav-item">
							<RouterLink to="/link1" class="nav-link">
								<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#layout"/></svg>
								Menu item 1
							</RouterLink>
						</li>
						<li class="nav-item">
							<RouterLink to="/link2" class="nav-link">
								<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#key"/></svg>
								Menu item 2
							</RouterLink>
						</li>
					</ul>

					<h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted text-uppercase">
						<span>Secondary menu</span>
					</h6>
					<ul class="nav flex-column">
						<li class="nav-item">
							<RouterLink :to="'/some/' + 'variable_here' + '/path'" class="nav-link">
								<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#file-text"/></svg>
								Item 1
							</RouterLink>
						</li>
					</ul>
				</div>
			</nav>
			
			<main class="col-md-9 ms-sm-auto col-lg-10 px-md-4">
				<div>
				<HomeView @login-success="refresh" @logout-success="refresh" />
			</div>
			</main>
		</div>
	</div>
</template>

<style>
</style>
