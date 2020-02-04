<template>
	<v-icon v-if="make" :title="model" v-model="attr">mdi-{{ camera_icon }}</v-icon>
</template>

<script>

export default {
	props: {
		make: {},
		model: {},
		attr: {},
	},
	computed: {
		camera_icon() {
			switch (this.make) {
				case 'SONY': return 'alpha'
				case 'Canon': return 'alpha-c'
				case 'iPhone': return 'apple'
			}
		},
	}
}
</script>