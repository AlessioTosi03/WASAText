<script>
export default {
	data: function() {
		return {
			errormsg: null,
			loading: false,
			some_data: null,
		}
	},
	methods: {
		async refresh() {
			this.loading = true;
			this.errormsg = null;
			try {
				let response = await this.$axios.get("/stream", {
					headers: { Authorization: "Bearer 1" }
				});
				this.conversations = response.data;
			} catch (e) {
				this.errormsg = e.toString();
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
	<div>
		<div
			class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
			<h1 class="h2">Home page</h1>
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
					<button type="button" class="btn btn-sm btn-outline-primary" @click="newItem">
						New
					</button>
				</div>
			</div>
		</div>

		<ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>
		<ul>
			<li v-for="(c, index) in conversations" :key="index">
				<!-- If it's a group conversation, show the group name -->
				<div v-if="c.conversation.type === 'group'">
				<strong>{{ c.group.name }}</strong> <!-- Group name -->
				</div>

				<!-- If it's a chat conversation, show the other participant's name -->
				<div v-else>
				<strong>{{ c.other_user }}</strong> <!-- Other user's name in chat -->
				</div>
			</li>
		</ul>
	</div>

</template>

<style>
</style>
