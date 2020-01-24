<template>
	<v-list :dense="photos.length > 10" class="photo-list" :class="{ ultradense: photos.length >= 20 }">
		<v-list-item v-for="(p,i) in photos" :key="'sel'+i">
			<v-list-item-content>
				<v-list-item-title><name :name="p.name" :exif="p.exif" /></v-list-item-title>
				<v-list-item-subtitle v-if="photos.length < 20">
					<date :date="p.exif.DateTimeOriginal" :tz="p.exif.TimeZone" :offset="p.exif.OffsetTime" />
					<rating v-if="photos.length < 15" :readonly="true" :value="p.meta.rating" />
				</v-list-item-subtitle>
			</v-list-item-content>
		</v-list-item>
	</v-list>
</template>


<script>
import Name from '~/components/info/name'
import Date from '~/components/info/date'
import Rating from '~/components/rating'


export default {
	props: {
		photos: {}
	},
	components: { Name, Date, Rating }
}
</script>

<style>
.photo-list.v-list--dense .v-list-item__content {
	padding: 4px 0;
}

.photo-list.ultradense .v-list-item {
	min-height: 20px;
}
</style>