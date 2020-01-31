<template>
	<div class="path-crumbs" v-if="path.length > 0">
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
	props: {
		path: {}
	},
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
	},
	methods: {
		click(item) {
			if (item.last) {
				return null
			}
			if (item.i === 0) {
				this.$emit('clear')
				return
			}
			while (this.path[this.path.length-1] !== item.text) {
				this.$emit('pop')
			}
		},
	}
}
</script>

<style>

</style>