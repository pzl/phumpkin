<template>
	<v-sheet>
		<v-row v-for="(s,i) in sliders" :key="i" class="simple-slider">
			<v-slider 
				dense height="4"
				readonly
				:label="s.title"
				:min="s.min" :max="s.max" step="0.01"
				:value="s.value"
			 />
			 <span class="caption" style="height: 4px; margin-top: -7px">{{ s.format ? s.format(s.value) : s.value }}</span>
		</v-row>
	</v-sheet>
</template>


<script>

export default {
	props: {
		sliders: {},
	},
}
</script>

<style>
.simple-slider .v-label {
	font-size: 13px;
}

.simple-slider .v-slider__thumb {
	height: 7px;
	width: 7px;
}
</style>