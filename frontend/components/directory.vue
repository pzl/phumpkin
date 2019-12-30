<template>
	<v-card class="dir" @click="">
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