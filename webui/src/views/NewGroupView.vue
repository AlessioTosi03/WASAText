<script>
export default {
	data: function() {
		return {
			errormsg: null,
			loading: false,
		}
	},
	methods: {

	},
}
</script>

<template>
	<div>
		<div
			class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
			<h1 class="h2">New Group</h1>
		</div>

		<ErrorMsg v-if="errormsg" :msg="errormsg"></ErrorMsg>

	</div>

</template>

<style>
</style>