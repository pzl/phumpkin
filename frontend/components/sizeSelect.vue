<template>
	<v-menu v-model="menu" :position-x="x" :position-y="y" :absolute="!!(x && y)" offset-y >
		<v-list>
			<v-list-item v-for="(s,i) in sizes" :key="i" :href="s.url" target="_blank" two-line dense>
				<v-list-item-content>
					<v-list-item-title>{{ s.name }}</v-list-item-title>
					<v-list-item-subtitle v-if="!!(s.width && s.height)">{{ s.width }}x{{ s.height }}</v-list-item-subtitle>
				</v-list-item-content>
			</v-list-item>
		</v-list>
	</v-menu>
</template>

<script>

export default {
	props: {
		x: {},
		y: {},
		thumbs: {},
		value: {},
	},
	data() {
		return {

		}
	},
	computed: {
		menu: {
			get() { return this.value },
			set(val) { this.$emit('input', val) },
		},
		sizes() {
			// multiple images selected
			if (Array.isArray(this.thumbs) && this.thumbs.length > 1) {
				return Object.keys(this.thumbs[0])
						.sort((a,b) => this.thumbs[0][a].width - this.thumbs[0][b].width)
						.map(s => {
							return { name: s }
						})
			}

			let t = this.thumbs
			if (Array.isArray(t) && t.length == 1) {
				t = t[0]
			}

			return Object.keys(t).map(s => {
				return {
					name: s,
					...t[s]
				}
			}).sort((a,b) => a.width - b.width)
		},
	}
}
</script>