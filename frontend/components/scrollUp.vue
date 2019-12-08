<template>
	<v-fab-transition>
		<v-btn v-show="show" v-scroll="onScroll" aria-label="Scroll to Top" bottom right fixed dark title="Scroll to Top" large fab @click="toTop" color="red" style="bottom: 60px">
			<v-icon>mdi-chevron-up</v-icon>
		</v-btn>
	</v-fab-transition>
</template>

<script>
export default {
	data() {
		return {
			show: false
		}
	},
	methods: {
		onScroll() {
			if (typeof window === 'undefined') {
				return
			}

			const top = ( window.pageYOffset || document.documentElement.offsetTop || 0)
			this.show = top > 300
		},
		toTop() {
			this.$router.push({ hash: '' })
			this.$vuetify.goTo(0)
		}
	}
}
</script>

<style>

</style>