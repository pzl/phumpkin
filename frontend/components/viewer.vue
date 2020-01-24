<template>
	<v-overlay :value="true" z-index="99">
		<v-sheet height="100%" width="100%">
			<v-btn icon @click="close"><v-icon>mdi-close</v-icon></v-btn>
			<v-img v-if="photos.length === 1" :height="photos[0].thumbs['large'].height" :width="photos[0].thumbs['large'].width" :lazy-src="photos[0].thumbs['x-small'].url+'?purpose=lazysrc'" :src="photos[0].thumbs['large'].url+'?purpose=viewer'" />
			<v-carousel v-else v-model="position" show-arrows-on-hover height="100%" style="width: 100%">
				<v-carousel-item v-for="(img,i) in photos" :key="i">
					<v-img :height="img.thumbs['large'].height" :width="img.thumbs['large'].width" :lazy-src="img.thumbs['x-small'].url+'?purpose=lazysrc'" :src="img.thumbs['large'].url+'?purpose=viewer'" />
				</v-carousel-item>
			</v-carousel>
		</v-sheet>
	</v-overlay>
</template>

<script>

export default {
	props: {
		photos: {},
	},
	data() {
		return {
			position: 0,
		}
	},
	methods: {
		close() {
			this.$emit('close')
		},
		click(e) {
			if (e.target.classList.contains('v-overlay__scrim')) { // clicked off content
				this.close()
			}
		},
		keys(ev) {
			switch (ev.keyCode) {
				case 27: // esc
					this.close()
					break
				case 37: // left
					this.position--
					break
				case 39: // right
					this.position++
					break
			}
		},
	},
	mounted() {
		document.addEventListener('mousedown', this.click)
		window.addEventListener('keydown', this.keys)
	},
	destroyed() {
		document.removeEventListener('mousedown', this.click)
		window.addEventListener('keydown', this.keys)
	}
}
</script>