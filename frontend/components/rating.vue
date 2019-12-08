<template>
	<div class="d-flex" @mouseover="hover = true" @mouseleave="hover = false">
		<v-icon
			v-if="hover || value === -1"
			@mouseover="hover_reject = true"
			@mouseleave="hover_reject = false"
			:color="value === -1 || hover_reject ? 'red' : 'grey lighten-1'"
			@click="$emit('input', -1)" small>mdi-cancel</v-icon>
		<v-rating v-if="hover || value >= 0" clearable dense color="yellow darken-4" small hover :value="rating" @input="$emit('input', $event)"></v-rating>
	</div>
</template>

<script>

export default {
	props: {
		value: {},
	},
	data() {
		return {
			hover: false,
			hover_reject: false,
		}
	},
	computed: {
		rating() {
			return this.hover_reject ? 0 : this.value
		}
	}
}
</script>