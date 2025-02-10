<script>
export default {
	data: function() {
		return {
			errormsg: null,
			loading: false,
			isLoggedIn: false,
            url: "",
            file: null,
		}
	},
	methods: {
        async setUsername() {
            this.errormsg = null;
            this.loading = true;
            console.log("Button clicked")
            const newUsername = prompt("Enter new username:");
            if (!newUsername) {
                return;
            }
            const token = localStorage.getItem("token");
            if (!token) {
                this.errormsg = "No token found. Please log in.";
                return;
            }
            this.url = "/users/" + localStorage.getItem("token");
            try {
                await this.$axios.put(this.url + "/name", { username: newUsername }, { 
                    headers: { Authorization: `Bearer ${token}` }
                });
                localStorage.setItem("username", newUsername);
                this.username = newUsername;
                this.$emit("username-changed", newUsername);
            } catch (e) {
                this.errormsg = e.toString();
            }
        },
        async setUserPicture(){
            this.loading = true;
			this.errormsg = null;

			try {
				const token = localStorage.getItem("token");
				if (!token) {
					this.errormsg = "No token found. Please log in.";
					this.loading = false;
					return;
				}
                this.url = "/users/" + localStorage.getItem("token") + "/photo";
                const formData = new FormData();
                if(this.file){
                    formData.append("photo", this.file);
                }
				// Send request
				let response = await this.$axios.put(this.url, formData, {
					headers: {
						Authorization: `Bearer ${token}`,
						"Content-Type": "multipart/form-data"
					}
				});
                console.log(response.data);
                localStorage.setItem("propic", response.data.photo_path);
				this.$emit("set-propic");
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
		}
    },

	mounted() {
		this.id = localStorage.getItem("token");
		this.username = localStorage.getItem("username");
		this.propic = localStorage.getItem("propic");
	}
}
</script>

<template>
	<div id="user-settings">
        <h1 class="border-bottom">User Settings</h1>
        <br>
        <br>
		<ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>
        <button type="button" class="user btn btn-sm btn-outline-primary" @click.prevent="setUsername">
            Change Username
        </button>
        <br>
        <br>
        <input type="file" @change="onFileChange" class="form-control-file">
        <button type="button" class="user btn btn-sm btn-outline-primary" @click.prevent="setUserPicture">
            Change User Picture
        </button>
	</div>

</template>

<style>
</style>