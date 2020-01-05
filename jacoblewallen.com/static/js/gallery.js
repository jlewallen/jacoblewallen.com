$(document).ready(function() {
	const items = [];

	$('figure').each(function() {
		if ($(this).attr('class') == 'no-photoswipe') {
			return true;
		}

		const $figure = $(this),
			$a = $figure.find('a'),
			$img = $figure.find('img'),
			$src = $a.attr('href'),
			$title = $img.attr('alt'),
			$msrc = $img.attr('src');

		if (!$a.data('size')) {
			console.log("no dimensions for " + $src);
			return true;
	   	}

		const $size = $a.data('size').split('x');
		const item = {
			src: $src,
			w: $size[0],
			h: $size[1],
			title: $title,
			msrc: $msrc
		};

		const index = items.length;

		items.push(item);

		$figure.on('click', function(event) {
			event.preventDefault();

			const $pswp = $('.pswp')[0];

			const options = {
				index: index,
				bgOpacity: 0.8,
				showHideOpacity: true
			};

			console.log("click", $pswp, items, options);

			const gallery = new PhotoSwipe($pswp, PhotoSwipeUI_Default, items, options);

			gallery.init();
		});
	});
});
