<template>
	<v-chip-group column>
		<v-chip v-for="(t,i) in tags" :key="i" :dark="dark" small>{{ t }}</v-chip>
	</v-chip-group>
</template>

<script>

export default {
	props: {
		tags: {},
		dark: {
			type: Boolean,
			default: true,
		},
	},
	data() {
		return {

		}
	}
}
</script>