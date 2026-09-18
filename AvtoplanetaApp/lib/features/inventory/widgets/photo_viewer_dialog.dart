import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:share_plus/share_plus.dart';
import '../../../core/api/api_client.dart';

class PhotoViewerDialog extends StatefulWidget {
  final List<String> photos;
  final int initialIndex;
  final String title;

  const PhotoViewerDialog({
    super.key,
    required this.photos,
    this.initialIndex = 0,
    this.title = 'Фото детали',
  });

  static void show(
    BuildContext context, {
    required List<String> photos,
    int initialIndex = 0,
    String title = 'Фото детали',
  }) {
    if (photos.isEmpty) return;
    Navigator.of(context).push(
      PageRouteBuilder(
        opaque: false,
        barrierDismissible: true,
        pageBuilder: (ctx, anim1, anim2) => PhotoViewerDialog(
          photos: photos,
          initialIndex: initialIndex,
          title: title,
        ),
      ),
    );
  }

  @override
  State<PhotoViewerDialog> createState() => _PhotoViewerDialogState();
}

class _PhotoViewerDialogState extends State<PhotoViewerDialog> {
  late PageController _pageController;
  late int _currentIndex;
  final TransformationController _transformationController =
      TransformationController();

  @override
  void initState() {
    super.initState();
    _currentIndex = widget.initialIndex.clamp(0, widget.photos.length - 1);
    _pageController = PageController(initialPage: _currentIndex);
  }

  @override
  void dispose() {
    _pageController.dispose();
    _transformationController.dispose();
    super.dispose();
  }

  String get _currentPhotoUrl =>
      apiClient.resolveUrl(widget.photos[_currentIndex]);

  void _copyLink() {
    final url = _currentPhotoUrl;
    Clipboard.setData(ClipboardData(text: url));
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Ссылка на фото скопирована в буфер обмена'),
        duration: Duration(seconds: 2),
      ),
    );
  }

  Future<void> _sharePhoto() async {
    final url = _currentPhotoUrl;
    try {
      final uri = Uri.tryParse(url);
      if (uri != null) {
        await SharePlus.instance.share(ShareParams(uri: uri, subject: widget.title));
      } else {
        _copyLink();
      }
    } catch (e) {
      _copyLink();
    }
  }

  void _resetZoom() {
    _transformationController.value = Matrix4.identity();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black.withValues(alpha: 0.96),
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        foregroundColor: Colors.white,
        elevation: 0,
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              widget.title,
              style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
            if (widget.photos.length > 1)
              Text(
                '${_currentIndex + 1} из ${widget.photos.length}',
                style: const TextStyle(fontSize: 12, color: Colors.white70),
              ),
          ],
        ),
        actions: [
          IconButton(
            tooltip: 'Сбросить масштаб',
            icon: const Icon(Icons.zoom_out_map_rounded),
            onPressed: _resetZoom,
          ),
          IconButton(
            tooltip: 'Скопировать ссылку',
            icon: const Icon(Icons.copy_rounded),
            onPressed: _copyLink,
          ),
          IconButton(
            tooltip: 'Поделиться',
            icon: const Icon(Icons.share_rounded),
            onPressed: _sharePhoto,
          ),
          IconButton(
            tooltip: 'Закрыть',
            icon: const Icon(Icons.close_rounded),
            onPressed: () => Navigator.of(context).pop(),
          ),
        ],
      ),
      body: PageView.builder(
        controller: _pageController,
        itemCount: widget.photos.length,
        onPageChanged: (idx) {
          setState(() {
            _currentIndex = idx;
            _resetZoom();
          });
        },
        itemBuilder: (ctx, i) {
          final url = apiClient.resolveUrl(widget.photos[i]);
          return InteractiveViewer(
            transformationController: _transformationController,
            minScale: 1.0,
            maxScale: 5.0,
            child: Center(
              child: CachedNetworkImage(
                imageUrl: url,
                fit: BoxFit.contain,
                placeholder: (ctx, url) => const Center(
                  child: CircularProgressIndicator(color: Colors.white),
                ),
                errorWidget: (ctx, url, error) => const Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(Icons.broken_image_outlined,
                          size: 64, color: Colors.white38),
                      SizedBox(height: 8),
                      Text('Не удалось загрузить фото',
                          style: TextStyle(color: Colors.white60)),
                    ],
                  ),
                ),
              ),
            ),
          );
        },
      ),
    );
  }
}
