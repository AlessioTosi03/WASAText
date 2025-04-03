<script setup>
import { RouterLink, RouterView } from 'vue-router'
import { getUrl } from './services/axios';
</script>
<script>
export default {
	data: function() {
		return {
			isLoggedIn: false,
			errormsg: null,
			loading: false,
			conversations: [],
			username: "",
			photo: "",
			user_url: "/users/",
			users: [],
			query: "",
			showDiv: false,
			showResults: false,
			propic : localStorage.getItem('propic')
		}
	},
	methods: {
		handleLoginSuccess() {
			this.refresh();  // Call the refresh method
			this.showDiv = true;  // Change the variable (set to false in this case)
			this.isLoggedIn= true;
		},
		handleLogoutSuccess() {
			this.refresh();  // Call the refresh method
			this.showDiv = false;  // Change the variable (set to false in this case)
			this.isLoggedIn= false;
		},
		async refresh() {
			this.loading = true;
			this.errormsg = null;

			const token = localStorage.getItem("token");  // Get the user_id from storage

			if (!token) {
				this.isLoggedIn = false;  // User is NOT logged in
				this.conversations = [];
				this.loading = false;
				return;
			}
			this.username = localStorage.getItem("username");
			this.photo = localStorage.getItem("propic");
			this.propic = localStorage.getItem('propic');
			this.user_url = "/users/" + localStorage.getItem("token");
			try {
				let response = await this.$axios.get("/stream", {
					headers: { Authorization: `Bearer ${token}` }
				});
				
				// If the request is successful, update conversations
				this.conversations = response.data;
				this.isLoggedIn = true;  // User is logged in

				await this.fetchUsers();
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
		async authError() {
			this.isLoggedIn = false;
			this.refresh();
			this.errormsg = "Invalid token. Please log in again.";
		},
		async fetchUsers() {
			this.loading = true;
			this.errormsg = null;

			try {
				let response = await this.$axios.get("/users");  // API Go
				const currentUser = localStorage.getItem("username");

				// Filtra l'utente corrente dalla lista
				this.users = response.data.filter(user => user !== currentUser);
			} catch (e) {
				this.errormsg = `Error: ${e.response ? e.response.data : e.toString()}`;
			}

			this.loading = false;
		},
		async createChat(username) {
			console.log("Creating chat with", username);
			this.query = username;
			this.loading = true;
			this.errormsg = null;
			try {
				const token = localStorage.getItem("token");
				if (!token) {
					this.errormsg = "No token found. Please log in.";
					this.loading = false;
					return;
				}
				
				await this.$axios.post("/newChat", {
					chatter: username,
					id: Number(localStorage.getItem("token"))
				}, {
					headers: { Authorization: `Bearer ${token}` }
				});
				this.$router.replace(`/`);
				this.$emit("new-chat");
				await this.refresh()
			} catch (e) {
				if (e.response && e.response.status === 401) {
					localStorage.removeItem("token");
					this.errormsg = "Invalid token. Please log in again.";
				} else {
					this.errormsg = `Error: ${e.response ? e.response.data : e.toString()}`;
				}
			}
		},
		filteredUsers() {
			if (!this.query) return [];
			return this.users.filter(user =>
				user.toLowerCase().includes(this.query.toLowerCase())
			);
		},
		showSearchResults() {
			this.showResults = true;
		},
		hideResults() {
			setTimeout(() => {  // Delay to ensure the click on the result item is registered
				this.showResults = false;
			}, 100);
		},
	},
	
	mounted() {
	this.propic = localStorage.getItem('propic')
		this.refresh()
		if (localStorage.getItem('token')) {
			this.showDiv = true;
			this.fetchUsers();
		}
	}
}
</script>

<template>

	<header class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow">
		<RouterLink class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6" v-if="isLoggedIn && (user_url && propic !== 'default')" :to="user_url">
			<img :src="getUrl(photo)" width="40" height="40" class="profile-pic" id="profile-pic">
			<p class="profile-name">{{ username }}</p>
		</RouterLink>
		<div class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6" v-else>
			<img src="/photos/default.png" width="40" height="40" class="profile-pic" id="profile-pic">
			<RouterLink v-if="isLoggedIn"  :to="user_url">
				<p class="profile-name">{{ username }}</p>
			</RouterLink>
			<p class="profile-name" v-else>Guest</p>
		</div>
		<button class="navbar-toggler position-absolute d-md-none collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#sidebarMenu" aria-controls="sidebarMenu" aria-expanded="false" aria-label="Toggle navigation">
			<span class="navbar-toggler-icon"></span>
		</button>
		<div class="search-container" v-if="showDiv">
			<input 
				v-model="query" 
				@focus="showResults = true"
				class="form-control form-control-dark w-80"
				type="text"
				placeholder="Search for users..."
				aria-label="Search"
			/>
			<div v-if="showResults && filteredUsers().length" class="dropdown">
				<div v-for="user in filteredUsers()" :key="user" @click="createChat(user)">
					{{ user }}
				</div>
			</div>
		</div>
	</header>

	<div class="container-fluid">
		<div class="row">
			<nav id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse">
				<div class="position-sticky pt-3 sidebar-sticky">

					<ul class="nav flex-column">
						<li class="nav-item" id="home-container">
							<RouterLink to="/" class="nav-link">
								<svg class="feather" id="home-svg"><use href="/feather-sprite-v4.29.0.svg#home"/></svg>
								Home
							</RouterLink>
						</li>
					</ul>

					<h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted text-uppercase">
						<span>Conversations</span>
					</h6>
					<ul class = "conversations-list">
						<li v-for="(c, index) in conversations" :key="index" id="conv-container" class="nav-item">
							<router-link  :to="'/chat/' + c.conversation.id + '/messages'">
									<!-- If it's a group conversation, show the group name -->
									<div v-if="c.conversation.type === 'group'" class = "conversation-item">
										<img :src="getUrl(c.group.photo)" v-if="isLoggedIn" width="40" height="40" class="conversation" id= "conversation-photo">
										<p class="conversation" id="conversation-name">{{ c.group.name }}</p> <!-- Group name -->
									</div>

									<!-- If it's a chat conversation, show the other participant's name -->
									<div v-else class = "conversation-item">
										<img :src="getUrl(c.other_user.profile_pic)" width="40" height="40" class="conversation" id="conversation-photo">
										<p class="conversation" id="conversation-name">{{ c.other_user.username }}</p>
									</div>
							</router-link>
						</li>
					</ul>
				</div>
			</nav>
			
			<main class="col-md-9 ms-sm-auto col-lg-10 px-md-4">
				<div id="content" >
					<ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>
				<RouterView @login-success="handleLoginSuccess" @logout-success="handleLogoutSuccess" @new-chat="refresh" @new-group="refresh" @conversation-loaded="refresh" @group-left="refresh" @group-added="refresh" @username-changed="refresh" @set-propic="refresh" @group-pic-updated="refresh" @group-name-updated="refresh" @message-forwarded="refresh" @message-deleted="refresh" @auth-error="authError"></RouterView>
			</div>
			</main>
		</div>
	</div>
</template>

<style>
	#content{
		position: relative;
		top: 20px;
	}
</style>
