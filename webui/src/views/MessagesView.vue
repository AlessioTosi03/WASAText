<script setup>
    import 'emoji-picker-element';
    import { getUrl } from '../services/axios';
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
            message: 0,
            url: "",
            text: "",
            file: null,
            showPopup: false, // Controls visibility of the popup
            showPopup2: false, // Controls visibility of the popup
            conversations: [],
            isLoggedIn: false,
            activePicker: null,  // Store the current picker ID (1, 2, etc.)
            selectedEmoji: '',
            myReaction: '',
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
                this.conversations = [];
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
                response = await this.$axios.get("/stream", {
					headers: { Authorization: `Bearer ${token}` }
				});
				
				// If the request is successful, update conversations
				this.conversations = response.data;
				this.isLoggedIn = true;  // User is logged in

                if (convData.conversation.type === "group") {
                    this.group = convData.group;
                } else {
                    this.chatter = convData.chatter;
                }
                this.conversation = convData.conversation;
                this.messages = convData.messages;
                if (this.messages === null) {
                    this.messages = [];
                }
                else {
                    for (let message of this.messages) {
                        if (!message.reaction) {
                            message.reaction = await this.getMyReaction(message.id);
                            console.log(message)
                        }
                    }
                }

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
                this.url = `/chat/${this.$route.params.conversation_id}/messages`;
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
                this.loading = false;
                this.$emit("auth-error");
                this.$router.push("/");
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
                    this.$emit("auth-error");
                    this.$router.push("/");
                } else {
                    this.errormsg = `Error: ${e.response ? e.response.data : e.toString()}`;
                }
            }

            finally {
                this.loading = false;
            }
        },
        async forwardMessage(conversation_id, message_id){
            console.log("Forwarding message", message_id, "to conversation", conversation_id);
            this.loading = true;
            this.errormsg = null;
            const token = localStorage.getItem("token");
            if (!token) {
                this.errormsg = "No token found. Please log in.";
                this.loading = false;
                return;
            }
            if (!conversation_id) {
                this.loading = false;
                return; // Skip fetching conversation data
            }
            this.url = `${conversation_id}`;
            try {
                let response = this.$axios.post(`/chat/${this.url}/forward`, {message_id,conversation_id}, {
                    headers: { Authorization: `Bearer ${token}` }
                });

                alert("Message forwarded successfully!");
                this.refresh();
                this.$emit("message-forwarded");
            } catch (e) {
                if (e.response && e.response.status === 401) {
                    localStorage.removeItem("token");
                    this.errormsg = "Invalid token. Please log in again.";
                    this.$emit("auth-error");
                    this.$router.push("/");

                } else {
                    this.errormsg = `Error: ${e.response ? e.response.data : e.toString()}`;
                }
            }

            finally {
                this.loading = false;
            }
        },
        async deleteMessage(message_id){
            console.log("Deleting message", message_id);
            this.loading = true;
            this.errormsg = null;
            const token = localStorage.getItem("token");
            if (!token) {
                this.errormsg = "No token found. Please log in.";
                this.loading = false;
                return;
            }
            if (!message_id) {
                this.loading = false;
                return; // Skip fetching conversation data
            }
            const convId = this.$route.params.conversation_id;

            try {
                await this.$axios.delete(`/chat/${convId}/messages/${message_id}`, {
                    headers: { Authorization: `Bearer ${token}` }
                });
                this.refresh();
                this.$emit("message-deleted");
            } catch (e) {
                if (e.response && e.response.status === 401) {
                    localStorage.removeItem("token");
                    this.errormsg = "Invalid token. Please log in again.";
                    this.$emit("auth-error");
                    this.$router.push("/");
                } else {
                    this.errormsg = `Error: ${e.response ? e.response.data : e.toString()}`;
                }
            }

            finally {
                this.loading = false;
            }
        },
        togglePicker(key) {
            if (this.activePicker === key) {
                this.activePicker = null; // Close the picker if it's already active
            } else {
                this.activePicker = key; // Open the selected picker
            }
        },
        async commentMessage(emoji,message_id) {
            this.selectedEmoji = `${emoji}`;
            this.activePicker = null; // Close the picker after selecting an emoji
            console.log(this.selectedEmoji)
            this.loading = true;
            this.errormsg = null;
            const token = localStorage.getItem("token");
            if (!token) {
                this.errormsg = "No token found. Please log in.";
                this.loading = false;
                return;
            }

            try {
                await this.$axios.post(`${this.url}/${message_id}/reaction`, {selectedEmoji: this.selectedEmoji}, {
                    headers: { Authorization: `Bearer ${token}` }
                });
                this.refresh();
                this.$emit("emoji-added");
            } catch (e) {
                if (e.response && e.response.status === 401) {
                    localStorage.removeItem("token");
                    this.errormsg = "Invalid token. Please log in again.";
                    this.$emit("auth-error");
                    this.$router.push("/");
                } else {
                    this.errormsg = `Error: ${e.response ? e.response.data : e.toString()}`;
                }
            }

            finally {
                this.loading = false;
            }
        },
        async getMyReaction(message_id){
            console.log("Getting my reaction to message", message_id);
            this.loading = true;
            this.errormsg = null;
            const token = localStorage.getItem("token");
            if (!token) {
                this.errormsg = "No token found. Please log in.";
                this.loading = false;
                return;
            }
            if (!message_id) {
                this.loading = false;
                return; // Skip fetching conversation data
            }
            const convId = this.$route.params.conversation_id;

            try {
                let response = await this.$axios.get(`/chat/${convId}/messages/${message_id}/reaction`, {
                    headers: { Authorization: `Bearer ${token}` }
                });
                console.log(response.data);
                this.myReaction = response.data;
                return this.myReaction;
            } catch (e) {
                if (e.response && e.response.status === 401) {
                    localStorage.removeItem("token");
                    this.errormsg = "Invalid token. Please log in again.";
                    this.$emit("auth-error");
                    this.$router.push("/");
                } else {
                    this.errormsg = `Error: ${e.response ? e.response.data : e.toString()}`;
                }
            }

            finally {
                this.loading = false;
            }
        },
        async uncommentMessage(message_id){
            console.log("Uncommenting message", message_id);
            this.loading = true;
            this.errormsg = null;
            const token = localStorage.getItem("token");
            if (!token) {
                this.errormsg = "No token found. Please log in.";
                this.loading = false;
                return;
            }
            if (!message_id) {
                this.loading = false;
                return; // Skip fetching conversation data
            }
            const convId = this.$route.params.conversation_id;

            try {
                await this.$axios.delete(`/chat/${convId}/messages/${message_id}/reaction`, {
                    headers: { Authorization: `Bearer ${token}` }
                });
                this.refresh();
                this.$emit("reaction-deleted");
            } catch (e) {
                if (e.response && e.response.status === 401) {
                    localStorage.removeItem("token");
                    this.errormsg = "Invalid token. Please log in again.";
                    this.$emit("auth-error");
                    this.$router.push("/");
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
        },
    },
    mounted() {
        this.refresh()
        this.username = localStorage.getItem("username");
        setInterval(this.refresh, 6000);
    }

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
	<div class="conv-container">
		<div v-if="conversation.type === 'group'" style="z-index:1000;" class="conv d-flex justify-content-start flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
            <img :src = "getUrl(group.photo)" width="70" height="70" class="conversation" id="conversation-photo">
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
        <div v-else style="z-index:1000;" class="conv d-flex justify-content-start flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
            <img :src = "getUrl(chatter.profile_pic)" width="70" height="70" class="conversation" id="conversation-photo">
            <h2 class="conversation" id="conversation-title">{{ chatter.username }}</h2>
        </div>
        <div id="messages-list" >
            <ul>
                <li v-for="(m, index) in messages" :key="index" class="message-item">
                    {{ m.username }}: {{ m.text }} 
                    <button @click="showPopup2 = true" class="btn btn-sm" id="forward-message">
                        <svg class="feather" id="forward-svg"><use href="/feather-sprite-v4.29.0.svg#corner-up-right"/></svg>
                    </button>
                    <div v-if="showPopup2" class="popup-overlay" @click.self="showPopup2 = false">
                        <div class="popup-content" id="convs-popup">
                            <ul class = "conversations-list">
                                <li v-for="(c, index) in conversations" :key="index" id="conv-container" class="nav-item">
                                    <button @click="forwardMessage(c.conversation.id,m.id)" class="btn btn-sm" id="forward-conv">
                                        <!-- If it's a group conversation, show the group name -->
                                        <div v-if="c.conversation.type === 'group'"  id="forward-box">
                                            <img :src="getUrl(c.group.photo)" width="40" height="40" class="conversation" id= "conversation-photo">
                                            <p class="conversation" id="conversation-name">{{ c.group.name }}</p> <!-- Group name -->
                                        </div>

                                        <!-- If it's a chat conversation, show the other participant's name -->
                                        <div v-else  id="forward-box">
                                            <img :src="getUrl(c.other_user.profile_pic)" width="40" height="40" class="conversation" id="conversation-photo">
                                            <p class="conversation" id="conversation-name">{{ c.other_user.username }}</p>
                                        </div>
                                    </button>
                                </li>
                            </ul>
                        </div>
                    </div>
                    <button v-if="m.username === username" @click="deleteMessage(m.id)" class="btn btn-sm"  id="delete-message">
                        <svg class="feather" id="delete-svg"><use href="/feather-sprite-v4.29.0.svg#trash-2"/></svg>
                    </button>
                    {{ m.reaction }}
                    <button v-if="!m.reaction" @click="togglePicker(index)">😊</button>
                    <emoji-picker :key="index"
                    v-if="activePicker === index" 
                    @emoji-click="commentMessage($event.detail.unicode, m.id)" 
                    class="emoji-picker"
                    ></emoji-picker>
                    <button v-if="m.reaction" @click="uncommentMessage(m.id)" class="btn btn-sm"  id="delete-reaction">
                        Delete reaction
                    </button>
                    <p v-if="m.forwarded==1" id="forwarded-text" style="margin-left: 30%">
                        Forwarded
                    </p>
                    <div v-if="m.pic != 'files/'" id="pic-container">
                        <img :src="getUrl(m.pic)">
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
