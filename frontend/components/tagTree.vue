<template>
	<v-treeview
		hoverable
		open-on-click
		:open-all="openAll"
		:items="items"
	/>
</template>

<script>

export default {
	props: {
		tags: {
			type: Array,
			default: [],
		},
	},
	computed: {
		items() {
			return this.tags.map(x => {
				return this.organize(x.split('|'))
			})
		},
		length() {
			return this.items.length
		},
		openAll() {
			return this.length < 10
		}
	},
	methods: {
		organize(bits) {
			if (bits.length === 0) {
				return []
			}
			if (bits.length === 1) {
				return { name: bits[0] }
			}
			return {
				name: bits.shift(),
				children: [this.organize(bits)]
			}
		}
	}
}
</script>