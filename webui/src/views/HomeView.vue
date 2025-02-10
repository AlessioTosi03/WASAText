<script>
export default {
	data: function() {
		return {
			errormsg: null,
			loading: false,
			isLoggedIn: false,
			username: "",
			propic: ""
		}
	},
	methods: {
		async refresh() {
			this.loading = true;
			this.errormsg = null;

			const token = localStorage.getItem("token");

			if (!token) {
				this.isLoggedIn = false;
				this.loading = false;
				return;
			}

			try {
				await this.$axios.get("/stream", { 
					headers: { Authorization: `Bearer ${token}` }
				});

				this.isLoggedIn = true;
			} catch (e) {
				if (e.response && e.response.status === 401) {
					this.isLoggedIn = false;
					localStorage.removeItem("token");
				} else {
					this.errormsg = e.toString();
				}
			}

			this.loading = false;
		},

		newChat() {
			this.$router.push("/newChat");
			this.$emit("new-chat");
		},
		newGroup() {
			this.$router.push("/newGroup");
			this.$emit("new-group");
		},
		login() {
			this.$axios.post("/session", { username: this.username })
				.then(response => {
					console.log(response.data);
					if (response.data) {
						localStorage.setItem("token", response.data.id);  // Store the user_id in localStorage
						localStorage.setItem("username", response.data.username);
						localStorage.setItem("propic", response.data.profile_pic);
						this.refresh();  // Check login status
						// Emit login state to App.vue to notify it to refresh
						this.$emit("login-success");  // Emit an event to parent when login is successful
					} else {
						this.errormsg = "Invalid response from server";
					}
				})
				.catch(e => {
					this.errormsg = e.response ? e.response.data.message : "An error occurred";
				});
		},
		logout() {
			localStorage.removeItem("token");
			localStorage.removeItem("username");
			localStorage.removeItem("propic");
			this.isLoggedIn = false;
			this.$router.push("/");
			this.refresh();  // Check login status
			this.$emit("logout-success");  // Emit an event to parent when login is successful
		},
	},
	mounted() {
		this.refresh();
		this.username = localStorage.getItem("username");
		this.propic = localStorage.getItem("propic");
	}
}
</script>

<template>
	<div>
		<div
			class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
			<h1 class="h2">example</h1>
			<div class="btn-toolbar mb-2 mb-md-0">
				<div class="btn-group me-2">
					<button type="button" class="btn btn-sm btn-outline-secondary" @click="refresh">
						Refresh
					</button>
					<button type="button" class="btn btn-sm btn-outline-secondary" @click="exportList">
						Export
					</button>
				</div>
				<div class="btn-group me-2">
					<button type="button" class="btn btn-sm btn-outline-primary" @click="newChat">
						New Chat
					</button>
				</div>
				<div class="btn-group me-2">
					<button type="button" class="btn btn-sm btn-outline-primary" @click="newGroup">
						New Group
					</button>
				</div>
			</div>
		</div>

		<ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>

		<div v-if="isLoggedIn" class = "indicator">
			<h1 id="welcome">Welcome to WASAText</h1>
			<p id="login">You are logged in as {{ username }}</p>
			<button id="logout-button" @click="logout">Log Out</button>
		</div>
		<div v-else class = "indicator">
			<h1 id="welcome">Welcome to WASAText</h1>
			<p id="login">Please log in to continue</p>
			<input type="text" id="username" placeholder="Username" v-model="username">
			<button id="login-button" @click="login">Log In</button>
		</div>
	</div>

</template>

<style>
</style>
