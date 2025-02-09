<script>
export default {
	data() {
		return {
			errormsg: null,
			loading: false,
			name: "",
			file: null
		};
	},
	methods: {
		async createGroup() {
			this.loading = true;
			this.errormsg = null;

			try {
				const token = localStorage.getItem("token");
				if (!token) {
					this.errormsg = "No token found. Please log in.";
					this.loading = false;
					return;
				}

				// Prepare FormData
				const formData = new FormData();
				formData.append("name", this.name);
				formData.append("id", Number(localStorage.getItem("token"))); // Convert to number
				if (this.file) {
					formData.append("photo", this.file); // Append the file
				}

				// Send request
				await this.$axios.post("/newGroup", formData, {
					headers: {
						Authorization: `Bearer ${token}`,
						"Content-Type": "multipart/form-data"
					}
				});

				this.$router.replace(`/`);
				this.$emit("new-group");
			} catch (e) {
				if (e.response && e.response.status === 401) {
					localStorage.removeItem("token");
					this.errormsg = "Invalid token. Please log in again.";
				} else {
					this.errormsg = `Error: ${e.response ? e.response.data : e.toString()}`;
				}
			} finally {
				this.loading = false;
			}
		},
		onFileChange(e) {
			this.file = e.target.files[0];
		},
	},
};
</script>

<template>
	<div>
		<div
			class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
			<h1 class="h2">New Group</h1>
		</div>

		<ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>

		<form @submit.prevent="createGroup">
			<p>Insert Group Name:</p>
			<input type="text" v-model="name" placeholder="Group Name" class="form-control" required />

			<p>Insert Group Picture:</p>
			<input type="file" id="idInput" @change="onFileChange" class="form-control" accept="image/*" />

			<br>
			<button type="submit" class="btn btn-primary" :disabled="loading">
				{{ loading ? "Creating..." : "Create Group" }}
			</button>
		</form>
	</div>
</template>


<style>
</style>