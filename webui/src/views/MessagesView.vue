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
                this.refresh();
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
            try {
                const token = localStorage.getItem("token");
                if (!token) {
                    this.errormsg = "No token found. Please log in.";
                    this.loading = false;
                    return;
                }
                await this.$axios.post(this.url, {
                    text: this.text
                }, {
                    headers: { Authorization: `Bearer ${token}` }
                });
                this.text = "";
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
            this.loading = false;
        }
	},
    mounted() {
        this.refresh()
    }
}
</script>

<template>
	<div class="conv-container">
		<div v-if="conversation.type === 'group'" class="d-flex justify-content-start flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
            <img :src = "group.photo" width="70" height="70" class="conversation" id="conversation-photo">
            <h2 class="conversation" id="conversation-title">{{ group.name }}</h2>
        </div>
        <!-- If it's a chat conversation, show the other participant's name -->
        <div v-else class="d-flex justify-content-start flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
            <img :src = "chatter.profile_pic" width="70" height="70" class="conversation" id="conversation-photo">
            <h2 class="conversation" id="conversation-title">{{ chatter.username }}</h2>
        </div>
        <div class="d-flex justify-content-start flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
            <ul>
                <li v-for="(m, index) in messages" :key="index" class="message-item">
                    {{ m.username }}: {{ m.text }}
                </li>
            </ul>
        </div>
        <form @submit.prevent="sendMessage" id="message-form" >
            <input type="text" v-model="text" id="message-input" placeholder="Type your message here">
            <button type="submit" class="btn btn-primary" :disabled="loading">Send</button>
        </form>
		<ErrorMsg v-if="errormsg" :msg="errormsg" ></ErrorMsg>
	</div>

</template>

<style>
</style>
