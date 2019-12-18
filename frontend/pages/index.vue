<template>
	<v-container fluid>
		<v-row :dense="view_dense">
			<v-col v-for="(img, i) in images" :key="i" :cols="view_size">
				<thumb v-bind="img" :index="i" @click="onClick(i, $event)"/>
			</v-col>
		</v-row>
	</v-container>
</template>

<script>
import Thumb from '~/components/thumb'
import { mapState, mapMutations, mapActions } from 'vuex'

export default {
	data() {
		return {

		}
	},
	computed: {
		...mapState('images', ['images', 'selected', 'loading']),
		...mapState('interface', ['view_size', 'view_dense'])
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
		...mapMutations('images', ['clearSelection']),
		...mapActions('images', ['toggleSelect', 'setSelection', 'addSelection', 'loadImages']),
	},
	mounted() {
		this.loadImages()
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

</style>