<template>
	<div class="tag-crumbs">
		<v-breadcrumbs
			v-for="(item,i) in items" :key="i"
			:items="item"
			divider="/"
			class="pa-0 text--primary"
		>
			<template v-slot:divider>
				<v-icon small class="primary--text">mdi-chevron-right</v-icon>
			</template>
		</v-breadcrumbs>
	</div>
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
			return this.tags.map(x => x.split('|').map(y => {
				return {
					text: y,
				}
			}))
		},
	},
}
</script>

<style>
.tag-crumbs .v-breadcrumbs li {
	font-size: 12px;
}
.tag-crumbs .v-breadcrumbs__divider {
	padding: 0 !important;
}
</style>