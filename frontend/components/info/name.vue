<template>
	<span>{{name}} <sup><camera-icon :make="exif.Make" :model="exif.Model" :attr="{ small: true }" /></sup></span>
</template>

<script>
import CameraIcon from '~/components/info/cameraIcon'

export default {
	props: {
		name: {},
		exif: {},
	},
	components: { CameraIcon },
}
</script>