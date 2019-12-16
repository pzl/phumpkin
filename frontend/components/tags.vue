<template>
	<div class="tags d-flex">
		<v-icon :dark="dark" x-small>mdi-tag</v-icon>
		<v-chip-group column>
			<v-chip v-for="(t,i) in tags" :key="i" :dark="dark" x-small>{{ t }}</v-chip>
		</v-chip-group>
	</div>
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