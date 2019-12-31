<template>
	<div class="path-crumbs" v-if="this.path.length > 0">
		<v-breadcrumbs :items="items" >
			<template v-slot:item="props">
				<v-breadcrumbs-item  :tag="props.item.last ?undefined:'a'" @click="click(props.item)">{{props.item.text}}</v-breadcrumbs-item>
			</template>
		</v-breadcrumbs>
	</div>
</template>

<script>
import { mapState, mapMutations, mapActions } from 'vuex'

export default {
	computed: {
		items() {
			return ['root'].concat(this.path).map((p,i) => {
				return {
					i: i,
					last: i === this.path.length,
					text: p,
				}
			})
		},
		...mapState('images', ['path'])
	},
	methods: {
		click(item) {
			if (item.last) {
				return null
			}
			if (item.i === 0) {
				this.clearPath()
				this.resetImages()
				return
			}
			while (this.path[this.path.length-1] !== item.text) {
				this.popPath()
			}
			this.resetImages()
		},
		...mapMutations('images', ['clearPath', 'popPath']),
		...mapActions('images', ['resetImages']),
	}
}
</script>

<style>

</style>