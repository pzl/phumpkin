<template>
	<span>{{name}} <sup><v-icon small v-if="exif.Make" :title="exif.Model">mdi-{{ camera_icon }}</v-icon></sup></span>
</template>

<script>

export default {
	props: {
		name: {},
		exif: {},
	},
	computed: {
		camera_icon() {
			if (!this.exif || !this.exif.Make) {
				return ''
			}
			switch (this.exif.Make) {
				case 'SONY': return 'alpha'
				case 'Canon': return 'alpha-c'
				case 'iPhone': return 'apple'
			}
		},
	},
}
</script>