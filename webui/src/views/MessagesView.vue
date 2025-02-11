<script setup>

</script>
<script>
export default {
	data: function() {
		return {
			errormsg: null,
			loading: false,
            conversationId: 0,
            conversation: {},
            chatter: "",
            group: {},
            messages: [],
            url: "",
            text: "",
            file: null,
            showPopup: false, // Controls visibility of the popup
		}
	},
	methods: {
        onFileChange(event) {
            this.file = event.target.files[0];
        },
        async refresh() {
			this.loading = true;
			this.errormsg = null;

			const token = localStorage.getItem("token");

			if (!token) {
				this.isLoggedIn = false;
				this.loading = false;
                this.errormsg = "No token found. Please log in.";
				return;
			}
            const convId = this.$route.params.conversation_id;
            if (!convId) {
                this.loading = false;
                return; // Skip fetching conversation data
            }
            this.url = `/chat/${convId}/messages`;
			try {
				let response = await this.$axios.get(this.url, { 
					headers: { Authorization: `Bearer ${token}` }
				});
                const convData = response.data;
                if (!convData.conversation) {
                    this.errormsg = "No conversation found.";
                    return;
                }

                if (convData.conversation.type === "group") {
                    this.group = convData.group;
                } else {
                    this.chatter = convData.chatter;
                }
                this.conversation = convData.conversation;
                this.messages = convData.messages;
				this.isLoggedIn = true;

                this.$emit("conversation-loaded", convData.conversation);
			} catch (e) {
				if (e.response && e.response.status === 401) {
					this.isLoggedIn = false;
					localStorage.removeItem("token");
                    this.errormsg = "Invalid token. Please log in again.";
				} else {
                    this.errormsg = `Error: ${e.response ? e.response.data : e.toString()}`;
				}
			}
			this.loading = false;
		},
        async sendMessage() {
            this.loading = true;
            this.errormsg = null;
            const formData = new FormData();
            formData.append("text", this.text); // Message text
            if (this.file) {
                formData.append("photo", this.file); // Attach the file
            }

            try {
                const token = localStorage.getItem("token");
                if (!token) {
                    this.errormsg = "No token found. Please log in.";
                    this.loading = false;
                    return;
                }
                if (!this.text && !this.file) {
                    this.errormsg = "Message or image required!";
                    this.loading = false;
                    return;
                }
                await this.$axios.post(this.url, formData, {
                    headers: { Authorization: `Bearer ${token}`,
                                "Content-Type": "multipart/form-data"
                     }
                });
                this.text = "";
                this.file = null;
                this.$refs.fileInput.value = "";
                this.refresh();
                this.$emit("message-sent");
            } catch (e) {
                if (e.response && e.response.status === 401) {
                    localStorage.removeItem("token");
                    this.errormsg = "Invalid token. Please log in again.";
                } else {
                    this.errormsg = `Error: ${e.response ? e.response.data : e.toString()}`;
                }
            }
            finally {
                this.loading = false;
            }
        },
        async leaveGroup(){
            this.loading = true;
            this.errormsg = null;
            const token = localStorage.getItem("token");
            if (!token) {
                this.errormsg = "No token found. Please log in.";
                this.loading = false;
                return;
            }
            const convId = this.$route.params.conversation_id;
            if (!convId) {
                this.loading = false;
                return; // Skip fetching conversation data
            }
            this.url = `${convId}`;
            try {
                await this.$axios.delete(`/chat/${this.url}`, {
                    headers: { Authorization: `Bearer ${token}` }
                });
                this.$router.push("/");
                this.$emit("group-left");
            } catch (e) {
                if (e.response && e.response.status === 401) {
                    localStorage.removeItem("token");
                    this.errormsg = "Invalid token. Please log in again.";
                } else {
                    this.errormsg = `Error: ${e.response ? e.response.data : e.toString()}`;
                }
            }
            finally {
                this.loading = false;
            }
        },
        async addToGroup(){
            this.loading = true;
            this.errormsg = null;
            const token = localStorage.getItem("token");
            if (!token) {
                this.errormsg = "No token found. Please log in.";
                this.loading = false;
                return;
            }
            const convId = this.$route.params.conversation_id;
            if (!convId) {
                this.loading = false;
                return; // Skip fetching conversation data
            }
            this.url = `${convId}`;
            const username = prompt("Enter your text:");
            if (username !== null) {
            console.log("User entered:", username);
            } else {
            console.log("User cancelled the dialog.");
            }
            try {
                await this.$axios.post(`/chat/${this.url}`, {username}, {
                    headers: { Authorization: `Bearer ${token}` }
                });
                alert("User added to the group successfully!");
                this.$router.push(`/chat/${convId}/messages`);
                this.$emit("group-added");
            } catch (e) {
                if (e.response && e.response.status === 401) {
                    localStorage.removeItem("token");
                    this.errormsg = "Invalid token. Please log in again.";
                } else {
                    this.errormsg = `Error: ${e.response ? e.response.data : e.toString()}`;
                }
            }

            finally {
                this.loading = false;
            }
        },
        async setGroupName(){
            this.loading = true;
            this.errormsg = null;
            const token = localStorage.getItem("token");
            if (!token) {
                this.errormsg = "No token found. Please log in.";
                this.loading = false;
                return;
            }
            const convId = this.$route.params.conversation_id;
            if (!convId) {
                this.loading = false;
                return; // Skip fetching conversation data
            }
            this.url = `${convId}`;
            try {
                await this.$axios.put(`/chat/${this.url}/name`, {name: this.text}, {
                    headers: { Authorization: `Bearer ${token}` }
                });
                alert("Group name updated successfully!");
                this.refresh();
                this.$emit("group-name-updated");
            } catch (e) {
                if (e.response && e.response.status === 401) {
                    localStorage.removeItem("token");
                    this.errormsg = "Invalid token. Please log in again.";
                } else {
                    this.errormsg = `Error: ${e.response ? e.response.data : e.toString()}`;
                }
            }

            finally {
                this.loading = false;
            }
        },
        async setGroupPic(){
            this.loading = true;
            this.errormsg = null;
            const token = localStorage.getItem("token");
            if (!token) {
                this.errormsg = "No token found. Please log in.";
                this.loading = false;
                return;
            }
            const convId = this.$route.params.conversation_id;
            if (!convId) {
                this.loading = false;
                return; // Skip fetching conversation data
            }
            this.url = `${convId}`;
            const formData = new FormData();
            if (this.file) {
                formData.append("photo", this.file); // Append the file
            }
            try {
                await this.$axios.put(`/chat/${this.url}/photo`, formData, {
                    headers: { Authorization: `Bearer ${token}`,
                                "Content-Type": "multipart/form-data"
                     }
                });
                alert("Group picture updated successfully!");
                this.refresh();
                this.$emit("group-pic-updated");
            } catch (e) {
                if (e.response && e.response.status === 401) {
                    localStorage.removeItem("token");
                    this.errormsg = "Invalid token. Please log in again.";
                } else {
                    this.errormsg = `Error: ${e.response ? e.response.data : e.toString()}`;
                }
            }

            finally {
                this.loading = false;
            }
        }
	},
    watch: {
        "$route.params.conversation_id": function (newId, oldId) {
            if (newId !== oldId) {
                this.refresh(); // Refresh when the conversation ID changes
            }
        }
    },
    mounted() {
        this.refresh()
    }
}
</script>

<template>
	<div class="conv-container">
		<div v-if="conversation.type === 'group'" class="conv d-flex justify-content-start flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
            <img :src = "group.photo" width="70" height="70" class="conversation" id="conversation-photo">
            <h2 class="conversation" id="conversation-title">{{ group.name }}</h2>
            <div id="popup-container">
                <!-- Button to show the popup -->
                <button @click="showPopup = true">Open Options</button>

                <!-- Popup -->
                <div v-if="showPopup" class="popup-overlay" @click.self="showPopup = false">
                <div class="popup-content">
                    <input type="text" v-model="text" id="set-group-name" placeholder="Type new group name here">
                    <button @click="setGroupName">Set Name</button>
                    <br><br>
                    <input type="file" ref="fileInput" @change="onFileChange" id="set-group-pic" accept="image/*">
                    <button @click="setGroupPic">Set Picture</button>
                    <br><br><br>
                    <button @click="showPopup = false">Close</button>
                </div>
                </div>
            </div>
            <button type="button" id="leave-group" class="btn btn-sm btn-outline-primary" @click="leaveGroup">
                Leave Group
            </button>
            <button type="button" id="add-to-group" class="btn btn-sm btn-outline-primary" @click="addToGroup">
                Add Member
            </button>
        </div>
        <!-- If it's a chat conversation, show the other participant's name -->
        <div v-else class="conv d-flex justify-content-start flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
            <img :src = "chatter.profile_pic" width="70" height="70" class="conversation" id="conversation-photo">
            <h2 class="conversation" id="conversation-title">{{ chatter.username }}</h2>
        </div>
        <div id="messages-list" >
            <ul>
                <li v-for="(m, index) in messages" :key="index" class="message-item">
                    {{ m.username }}: {{ m.text }}
                    <br>
                    <div v-if="m.pic" id="pic-container">
                        <img :src=m.pic class="message-pic">
                    </div>
                </li>
            </ul>
        </div>
        <ErrorMsg v-if="errormsg" :msg="errormsg" ></ErrorMsg>

        <form @submit.prevent="sendMessage" id="message-form" >
            <input type="text" v-model="text" id="message-input" placeholder="Type your message here">
            <input type="file" ref="fileInput" @change="onFileChange" id="message-pic" accept="image/*">
            <button type="submit" class="btn btn-primary" :disabled="loading">Send</button>
        </form>
	</div>

</template>

<style>
</style>
