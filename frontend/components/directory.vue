<template>
	<v-card class="dir" @click="$emit('click',$event)">
		<v-card-title><v-icon>mdi-folder</v-icon> {{name}}</v-card-title>
	</v-card>
</template>

<script>
export default {
	props: {
		name: {}
	},
}
</script>

<style>
.dir {
	margin: 10px 40px;
}
</style>