<template>
	<div class="d-flex flex-column pa-0">
		<slot name="before" />
		<v-row class="content-group pa-0 ma-0" justify="space-between" align="start">
			<thumb
				v-for="(img, i) in images" :key="i"
				v-bind="img"
				:index="i"
				@click="onClick(i, $event)"
			/>
			<div v-if="loading" class="my-12 d-flex" style="flex-basis: 100%">
				<v-progress-circular indeterminate class="mx-auto" color="deep-orange lighten-2" />
			</div>
			<span v-if="loadMore && !loading" class="ender" v-intersect="{ handler: intersect, options: { threshold: [0,1] }}"></span>
			<div v-if="err" class="my-12 d-flex" style="flex-basis: 100%">
				<v-btn icon class="mx-auto error--text" x-large @click="$emit('more')">
					<v-icon x-large>mdi-reload-alert</v-icon>
				</v-btn>
			</div>
		</v-row>
	</div>
</template>

<script>
import Thumb from '~/components/thumb'
import { mapState, mapMutations, mapActions } from 'vuex'

export default {
	props: {
		images: {},
	},
	data() {
		return {
		}
	},
	computed: {
		...mapState('images', ['selected', 'loading', 'err', 'loadMore']),
	},
	methods: {
		onClick(img, e) {
			e.preventDefault()

			// shift = add selection range
			if (e.shiftKey && this.selected.length) {
				let lastSel = this.selected[this.selected.length - 1]

				if (lastSel < img) {
					for (let i=lastSel; i<=img; i++) {
						this.addSelection(i)
					}
				} else {
					for (let i=img; i<=lastSel; i++) {
						this.addSelection(i)
					}
				}
				return
			}

			// ctrl = add image to current selection
			if (e.ctrlKey) {
				this.toggleSelect(img)
				return
			}

			// only reset selection if this is a newly clicked tile
			// this prevents losing a large range if clicking an already-selected one
			if (!this.selected.includes(img)) {
				this.setSelection(img)

			}
		},
		intersect(entries, observer, isIntersecting) {
			if (!isIntersecting) {
				return false
			}
			if (this.loading) {
				return false
			}
			if (this.err) {
				return false
			}
			console.log("scrolled to end. loading more")
			this.$emit('more')
		},
		...mapActions('images', ['toggleSelect', 'setSelection', 'addSelection']),
	},
	components: { Thumb }
}
</script>


<style>
.thumby {
	cursor: pointer;
}

.v-card--reveal {
	align-items: center;
  opacity: .5;
  bottom: 0;
  position: absolute;
  width: 100%;

  user-select: none; /* prevent name highlighting on shift-click */
}


.content-group {
	max-width: 100%;
}
</style>