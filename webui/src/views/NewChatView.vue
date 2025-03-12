<script>
export default {
	data: function() {
		return {
			errormsg: null,
			loading: false,
			id: 0,
			chatter: "",
		}
	},
	methods: {
		async createChat() {
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
					chatter: this.chatter,
					id: Number(localStorage.getItem("token"))
				}, {
					headers: { Authorization: `Bearer ${token}` }
				});
				this.$router.replace(`/`);
				this.$emit("new-chat");
			} catch (e) {
				if (e.response && e.response.status === 401) {
					localStorage.removeItem("token");
					this.errormsg = "Invalid token. Please log in again.";
				} else {
					this.errormsg = `Error: ${e.response ? e.response.data : e.toString()}`;
				}
			}
		},
	},
}
</script>

<template>
	<div class="fixed-top d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3">
        <h1 class="h2"></h1>
        <div class="btn-toolbar mb-2 mb-md-0">
            <div class="btn-group me-2">
                <button type="button" class="btn btn-sm btn-outline-secondary" @click="refresh">
                    Refresh
                </button>
                <button type="button" class="btn btn-sm btn-outline-secondary" @click="exportList">
                    Export
                </button>
            </div>
        </div>
    </div>
	<div>
		<div
			class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
			<h1 class="h2">New Chat</h1>
		</div>

		<ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>
		<form @submit.prevent="createChat">
			<p>Insert User to chat:</p><input type="text" v-model="chatter" placeholder="Username" class="form-control" />
			<br>
			<button type="submit" class="btn btn-primary" :disabled="loading">Create Chat</button>
		</form>
	</div>

</template>

<style>
</style>
