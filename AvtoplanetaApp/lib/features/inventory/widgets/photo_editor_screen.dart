import 'dart:io';
import 'dart:math' as math;
import 'dart:ui' as ui;
import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:dio/dio.dart';
import 'package:path_provider/path_provider.dart';
import 'package:share_plus/share_plus.dart';
import '../../../core/api/api_client.dart';
import '../providers/inventory_provider.dart';

enum EditorMode {
  view,
  crop,
  brush,
  blur,
  eraser,
}

enum StrokeType {
  brush,
  blur,
  eraser,
}

class EditorStroke {
  final List<Offset> points;
  final Color color;
  final double strokeWidth;
  final StrokeType type;

  EditorStroke({
    required this.points,
    required this.color,
    required this.strokeWidth,
    required this.type,
  });

  EditorStroke copy() => EditorStroke(
        points: List<Offset>.from(points),
        color: color,
        strokeWidth: strokeWidth,
        type: type,
      );
}

class PhotoEditorScreen extends ConsumerStatefulWidget {
  final File? initialFile;
  final String? initialUrl;
  final int? partId;
  final String? photoPathToReplace;

  const PhotoEditorScreen({
    super.key,
    this.initialFile,
    this.initialUrl,
    this.partId,
    this.photoPathToReplace,
  });

  static Future<dynamic> show(
    BuildContext context, {
    File? imageFile,
    String? imageUrl,
    int? partId,
    String? photoPathToReplace,
  }) {
    return Navigator.of(context).push(
      MaterialPageRoute(
        fullscreenDialog: true,
        builder: (_) => PhotoEditorScreen(
          initialFile: imageFile,
          initialUrl: imageUrl,
          partId: partId,
          photoPathToReplace: photoPathToReplace,
        ),
      ),
    );
  }

  @override
  ConsumerState<PhotoEditorScreen> createState() => _PhotoEditorScreenState();
}

class _PhotoEditorScreenState extends ConsumerState<PhotoEditorScreen> {
  bool _loading = true;
  String _statusText = 'Загрузка изображения...';

  ui.Image? _rawImage;
  ui.Image? _blurredImage;

  EditorMode _mode = EditorMode.crop;

  // Трансформации
  int _rotationQuarter = 0; // 0, 1, 2, 3 (* 90°)
  bool _flipH = false;
  bool _flipV = false;

  // Кадрирование (нормализованный Rect от 0.0 до 1.0)
  Rect _cropRect = const Rect.fromLTWH(0.0, 0.0, 1.0, 1.0);
  double? _aspectPreset; // null = free, 1.0 = square, 4/3, 16/9

  // Рисование
  Color _brushColor = const Color(0xFFEF4444); // Красный по умолчанию
  double _brushWidth = 8.0;
  double _blurWidth = 28.0;
  double _eraserWidth = 24.0;

  final List<EditorStroke> _strokes = [];
  EditorStroke? _currentStroke;

  // Undo / Redo стеки
  final List<List<EditorStroke>> _undoStack = [];
  final List<List<EditorStroke>> _redoStack = [];

  // Палитра цветов (в точности как на сайте в ImageCropper.tsx)
  final List<Color> _brushColors = const [
    Color(0xFFEF4444), // Красный
    Color(0xFFEAB308), // Желтый
    Color(0xFF22C55E), // Зеленый
    Color(0xFF3B82F6), // Синий
    Color(0xFFFFFFFF), // Белый
    Color(0xFF000000), // Черный
  ];

  @override
  void initState() {
    super.initState();
    _loadImage();
  }

  @override
  void dispose() {
    _rawImage?.dispose();
    _blurredImage?.dispose();
    super.dispose();
  }

  Future<void> _loadImage() async {
    setState(() {
      _loading = true;
      _statusText = 'Загрузка изображения...';
    });

    try {
      File file;
      if (widget.initialFile != null) {
        file = widget.initialFile!;
      } else if (widget.initialUrl != null) {
        final url = apiClient.resolveUrl(widget.initialUrl!);
        final tempDir = await getTemporaryDirectory();
        final ext = url.contains('.png') ? 'png' : 'jpg';
        file = File('${tempDir.path}/editor_src_${DateTime.now().millisecondsSinceEpoch}.$ext');
        await apiClient.dio.download(url, file.path);
      } else {
        throw Exception('Изображение не предоставлено');
      }

      setState(() => _statusText = 'Подготовка холста...');
      final bytes = await file.readAsBytes();
      final codec = await ui.instantiateImageCodec(bytes);
      final frame = await codec.getNextFrame();
      final raw = frame.image;

      // Создаем размытую версию для кисти цензуры/размытия
      final blurred = await _createBlurredImage(raw);

      if (mounted) {
        setState(() {
          _rawImage = raw;
          _blurredImage = blurred;
          _loading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() => _loading = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Ошибка загрузки фото: $e')),
        );
        Navigator.of(context).pop();
      }
    }
  }

  Future<ui.Image> _createBlurredImage(ui.Image source) async {
    final recorder = ui.PictureRecorder();
    final canvas = Canvas(recorder);
    final paint = Paint()
      ..imageFilter = ui.ImageFilter.blur(sigmaX: 18, sigmaY: 18, tileMode: TileMode.clamp);
    canvas.drawImage(source, Offset.zero, paint);
    final picture = recorder.endRecording();
    return await picture.toImage(source.width, source.height);
  }

  void _pushHistory() {
    _undoStack.add(_strokes.map((s) => s.copy()).toList());
    _redoStack.clear();
  }

  void _undo() {
    if (_undoStack.isEmpty) return;
    setState(() {
      _redoStack.add(_strokes.map((s) => s.copy()).toList());
      final prev = _undoStack.removeLast();
      _strokes
        ..clear()
        ..addAll(prev);
    });
  }

  void _redo() {
    if (_redoStack.isEmpty) return;
    setState(() {
      _undoStack.add(_strokes.map((s) => s.copy()).toList());
      final next = _redoStack.removeLast();
      _strokes
        ..clear()
        ..addAll(next);
    });
  }

  void _reset() {
    setState(() {
      _rotationQuarter = 0;
      _flipH = false;
      _flipV = false;
      _cropRect = const Rect.fromLTWH(0.0, 0.0, 1.0, 1.0);
      _aspectPreset = null;
      _strokes.clear();
      _undoStack.clear();
      _redoStack.clear();
    });
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Все изменения сброшены к оригиналу'), duration: Duration(seconds: 1)),
    );
  }

  void _rotateCw() {
    setState(() {
      _rotationQuarter = (_rotationQuarter + 1) % 4;
      _cropRect = const Rect.fromLTWH(0.0, 0.0, 1.0, 1.0);
    });
  }

  void _rotateCcw() {
    setState(() {
      _rotationQuarter = (_rotationQuarter + 3) % 4;
      _cropRect = const Rect.fromLTWH(0.0, 0.0, 1.0, 1.0);
    });
  }

  void _toggleFlipH() {
    setState(() => _flipH = !_flipH);
  }

  void _toggleFlipV() {
    setState(() => _flipV = !_flipV);
  }

  void _setAspect(double? aspect) {
    setState(() {
      _aspectPreset = aspect;
      if (aspect == null) {
        _cropRect = const Rect.fromLTWH(0.0, 0.0, 1.0, 1.0);
        return;
      }
      // Центрируем рамку с учетом выбранных пропорций
      final double targetAspect = aspect;
      double w = 1.0;
      double h = 1.0;
      if (targetAspect >= 1.0) {
        h = 1.0 / targetAspect;
      } else {
        w = targetAspect;
      }
      final l = (1.0 - w) / 2.0;
      final t = (1.0 - h) / 2.0;
      _cropRect = Rect.fromLTWH(l.clamp(0.0, 1.0), t.clamp(0.0, 1.0), w.clamp(0.1, 1.0), h.clamp(0.1, 1.0));
    });
  }

  // Рендеринг финального изображения со всеми правками
  Future<File> _renderExportFile() async {
    if (_rawImage == null) throw Exception('Изображение не загружено');

    final raw = _rawImage!;
    final int origW = raw.width;
    final int origH = raw.height;

    // Сначала рисуем полное изображение с размытием и маркерами
    final fullRecorder = ui.PictureRecorder();
    final fullCanvas = Canvas(fullRecorder);
    final fullBounds = Rect.fromLTWH(0, 0, origW.toDouble(), origH.toDouble());

    // 1. Исходная картинка
    fullCanvas.drawImage(raw, Offset.zero, Paint());

    // 2. Слой штрихов
    if (_strokes.isNotEmpty && _blurredImage != null) {
      fullCanvas.saveLayer(fullBounds, Paint());

      for (final stroke in _strokes) {
        if (stroke.points.isEmpty) continue;

        if (stroke.type == StrokeType.blur) {
          fullCanvas.saveLayer(fullBounds, Paint());
          final maskPaint = Paint()
            ..style = PaintingStyle.stroke
            ..strokeCap = StrokeCap.round
            ..strokeJoin = StrokeJoin.round
            ..strokeWidth = stroke.strokeWidth;

          if (stroke.points.length == 1) {
            fullCanvas.drawCircle(stroke.points.first, stroke.strokeWidth / 2, maskPaint..style = PaintingStyle.fill);
          } else {
            final path = Path()..moveTo(stroke.points.first.dx, stroke.points.first.dy);
            for (int i = 1; i < stroke.points.length; i++) {
              path.lineTo(stroke.points[i].dx, stroke.points[i].dy);
            }
            fullCanvas.drawPath(path, maskPaint);
          }

          fullCanvas.drawImage(_blurredImage!, Offset.zero, Paint()..blendMode = BlendMode.srcIn);
          fullCanvas.restore();
        } else if (stroke.type == StrokeType.brush) {
          final paint = Paint()
            ..color = stroke.color
            ..strokeWidth = stroke.strokeWidth
            ..style = PaintingStyle.stroke
            ..strokeCap = StrokeCap.round
            ..strokeJoin = StrokeJoin.round;

          if (stroke.points.length == 1) {
            fullCanvas.drawCircle(stroke.points.first, stroke.strokeWidth / 2, paint..style = PaintingStyle.fill);
          } else {
            final path = Path()..moveTo(stroke.points.first.dx, stroke.points.first.dy);
            for (int i = 1; i < stroke.points.length; i++) {
              path.lineTo(stroke.points[i].dx, stroke.points[i].dy);
            }
            fullCanvas.drawPath(path, paint);
          }
        } else if (stroke.type == StrokeType.eraser) {
          final erasePaint = Paint()
            ..blendMode = BlendMode.clear
            ..strokeWidth = stroke.strokeWidth
            ..style = PaintingStyle.stroke
            ..strokeCap = StrokeCap.round
            ..strokeJoin = StrokeJoin.round;

          if (stroke.points.length == 1) {
            fullCanvas.drawCircle(stroke.points.first, stroke.strokeWidth / 2, erasePaint..style = PaintingStyle.fill);
          } else {
            final path = Path()..moveTo(stroke.points.first.dx, stroke.points.first.dy);
            for (int i = 1; i < stroke.points.length; i++) {
              path.lineTo(stroke.points[i].dx, stroke.points[i].dy);
            }
            fullCanvas.drawPath(path, erasePaint);
          }
        }
      }

      fullCanvas.restore();
    }

    final fullPicture = fullRecorder.endRecording();
    final composedImage = await fullPicture.toImage(origW, origH);

    // 3. Применяем поворот, отражение и кадрирование
    final bool isRotated90 = _rotationQuarter % 2 != 0;
    final int rotatedW = isRotated90 ? origH : origW;
    final int rotatedH = isRotated90 ? origW : origH;

    // Пиксели кадрирования
    final int cropX = (_cropRect.left * rotatedW).round().clamp(0, rotatedW - 1);
    final int cropY = (_cropRect.top * rotatedH).round().clamp(0, rotatedH - 1);
    final int cropW = (_cropRect.width * rotatedW).round().clamp(10, rotatedW - cropX);
    final int cropH = (_cropRect.height * rotatedH).round().clamp(10, rotatedH - cropY);

    final finalRecorder = ui.PictureRecorder();
    final finalCanvas = Canvas(finalRecorder);

    // Смещаем начало координат для кропа
    finalCanvas.translate(-cropX.toDouble(), -cropY.toDouble());

    // Центр для вращения
    finalCanvas.save();
    finalCanvas.translate(rotatedW / 2, rotatedH / 2);
    finalCanvas.rotate(_rotationQuarter * math.pi / 2);
    finalCanvas.scale(_flipH ? -1.0 : 1.0, _flipV ? -1.0 : 1.0);
    finalCanvas.translate(-origW / 2, -origH / 2);

    finalCanvas.drawImage(composedImage, Offset.zero, Paint());
    finalCanvas.restore();

    final finalPicture = finalRecorder.endRecording();
    final finalImage = await finalPicture.toImage(cropW, cropH);

    final byteData = await finalImage.toByteData(format: ui.ImageByteFormat.png);
    final buffer = byteData!.buffer.asUint8List();

    final tempDir = await getTemporaryDirectory();
    final outPath = '${tempDir.path}/edited_${DateTime.now().millisecondsSinceEpoch}.png';
    final outFile = File(outPath);
    await outFile.writeAsBytes(buffer);

    composedImage.dispose();
    finalImage.dispose();

    return outFile;
  }

  // Скопировать результат в буфер без сохранения
  Future<void> _handleCopy() async {
    setState(() {
      _loading = true;
      _statusText = 'Копирование фото...';
    });

    try {
      final file = await _renderExportFile();
      // Системный буфер обмена поддерживает путь или текст
      await Clipboard.setData(ClipboardData(text: file.path));
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            backgroundColor: const Color(0xFF10B981),
            behavior: SnackBarBehavior.floating,
            content: Row(
              children: const [
                Icon(LucideIcons.circle_check, color: Colors.white),
                SizedBox(width: 8),
                Text('Отредактированное фото скопировано!', style: TextStyle(fontWeight: FontWeight.bold)),
              ],
            ),
            duration: const Duration(seconds: 3),
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Не удалось скопировать: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  // Поделиться фото
  Future<void> _handleShare() async {
    setState(() {
      _loading = true;
      _statusText = 'Подготовка к отправке...';
    });

    try {
      final file = await _renderExportFile();
      await SharePlus.instance.share(ShareParams(
        files: [XFile(file.path)],
        subject: 'Отредактированное фото запчасти',
      ));
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Ошибка отправки: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  // Сохранить отредактированное фото
  Future<void> _handleSave() async {
    setState(() {
      _loading = true;
      _statusText = 'Сохранение фото...';
    });

    try {
      final file = await _renderExportFile();

      // Если задан partId — сохраняем прямо на сервер
      if (widget.partId != null) {
        final partId = widget.partId!;
        final formData = FormData.fromMap({
          'photo': await MultipartFile.fromFile(file.path, filename: 'photo.png'),
        });

        // Загружаем новое фото
        await apiClient.dio.post('/api/uploadpartphoto/$partId', data: formData);

        // Если было старое фото для замены — удаляем его
        if (widget.photoPathToReplace != null) {
          try {
            await apiClient.dio.delete(
              '/api/deletepartphoto/$partId',
              queryParameters: {'photo': widget.photoPathToReplace},
            );
          } catch (delErr) {
            debugPrint('Non-critical: delete old photo error: $delErr');
          }
        }

        // Инвалидируем провайдеры деталей
        ref.invalidate(partProvider(partId));
        ref.invalidate(inventoryFilterProvider);

        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              backgroundColor: const Color(0xFF10B981),
              behavior: SnackBarBehavior.floating,
              content: Row(
                children: const [
                  Icon(LucideIcons.circle_check, color: Colors.white),
                  SizedBox(width: 8),
                  Text('Фото детали успешно сохранено!', style: TextStyle(fontWeight: FontWeight.bold)),
                ],
              ),
            ),
          );
          Navigator.of(context).pop(true);
        }
      } else {
        // Возвращаем экспортированный файл
        if (mounted) {
          Navigator.of(context).pop(file);
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Ошибка сохранения фото: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFF09090B),
      appBar: AppBar(
        backgroundColor: const Color(0xFF18181B),
        foregroundColor: Colors.white,
        elevation: 1,
        title: const Text(
          'Редактирование фото',
          style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
        ),
        actions: [
          IconButton(
            tooltip: 'Отменить (Undo)',
            icon: const Icon(LucideIcons.undo),
            onPressed: _undoStack.isNotEmpty ? _undo : null,
          ),
          IconButton(
            tooltip: 'Вернуть (Redo)',
            icon: const Icon(LucideIcons.redo),
            onPressed: _redoStack.isNotEmpty ? _redo : null,
          ),
          IconButton(
            tooltip: 'Скопировать результат в буфер без сохранения',
            icon: const Icon(LucideIcons.copy),
            onPressed: _rawImage != null && !_loading ? _handleCopy : null,
          ),
          IconButton(
            tooltip: 'Поделиться результатом',
            icon: const Icon(LucideIcons.share_2),
            onPressed: _rawImage != null && !_loading ? _handleShare : null,
          ),
          IconButton(
            tooltip: 'Сбросить к оригиналу',
            icon: const Icon(LucideIcons.rotate_ccw),
            onPressed: _rawImage != null && !_loading ? _reset : null,
          ),
        ],
      ),
      body: _loading
          ? Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const CircularProgressIndicator(color: Color(0xFF6366F1)),
                  const SizedBox(height: 16),
                  Text(_statusText, style: const TextStyle(color: Colors.white70)),
                ],
              ),
            )
          : Column(
              children: [
                // Центральный рабочий холст
                Expanded(
                  child: _buildCanvasViewport(),
                ),

                // Панель параметров выбранного режима
                _buildModeParametersPanel(),

                // Нижний тулбар переключения режимов
                _buildModeSelectorBar(),

                // Нижние кнопки Отмена / Сохранить
                _buildBottomActionButtons(),
              ],
            ),
    );
  }

  Widget _buildCanvasViewport() {
    return LayoutBuilder(
      builder: (context, constraints) {
        final canvasSize = Size(constraints.maxWidth, constraints.maxHeight);
        return GestureDetector(
          onPanStart: (details) => _onPanStart(details.localPosition, canvasSize),
          onPanUpdate: (details) => _onPanUpdate(details.localPosition, canvasSize),
          onPanEnd: (details) => _onPanEnd(),
          child: Container(
            color: const Color(0xFF09090B),
            width: double.infinity,
            height: double.infinity,
            child: CustomPaint(
              size: canvasSize,
              painter: _PhotoCanvasPainter(
                rawImage: _rawImage!,
                blurredImage: _blurredImage,
                rotationQuarter: _rotationQuarter,
                flipH: _flipH,
                flipV: _flipV,
                cropRect: _cropRect,
                mode: _mode,
                strokes: _strokes,
                currentStroke: _currentStroke,
              ),
            ),
          ),
        );
      },
    );
  }

  Offset? _screenToImageOffset(Offset localPos, Size canvasSize) {
    if (_rawImage == null) return null;
    final origW = _rawImage!.width.toDouble();
    final origH = _rawImage!.height.toDouble();

    final isRotated90 = _rotationQuarter % 2 != 0;
    final effW = isRotated90 ? origH : origW;
    final effH = isRotated90 ? origW : origH;

    final scale = math.min(canvasSize.width / effW, canvasSize.height / effH) * 0.94;
    final originX = (canvasSize.width - effW * scale) / 2;
    final originY = (canvasSize.height - effH * scale) / 2;

    final relX = (localPos.dx - originX) / scale;
    final relY = (localPos.dy - originY) / scale;

    if (relX < 0 || relX > effW || relY < 0 || relY > effH) {
      return null;
    }

    // Обратная трансформация к исходному rawImage
    double imgX = relX;
    double imgY = relY;

    if (_rotationQuarter == 1) {
      // 90° CW
      imgX = relY;
      imgY = origH - relX;
    } else if (_rotationQuarter == 2) {
      // 180°
      imgX = origW - relX;
      imgY = origH - relY;
    } else if (_rotationQuarter == 3) {
      // 270° CW
      imgX = origW - relY;
      imgY = relX;
    }

    if (_flipH) imgX = origW - imgX;
    if (_flipV) imgY = origH - imgY;

    return Offset(imgX.clamp(0.0, origW), imgY.clamp(0.0, origH));
  }

  void _onPanStart(Offset localPos, Size canvasSize) {
    if (_mode == EditorMode.brush || _mode == EditorMode.blur || _mode == EditorMode.eraser) {
      final imgPoint = _screenToImageOffset(localPos, canvasSize);
      if (imgPoint == null) return;

      _pushHistory();

      final type = _mode == EditorMode.brush
          ? StrokeType.brush
          : (_mode == EditorMode.blur ? StrokeType.blur : StrokeType.eraser);

      final width = _mode == EditorMode.brush
          ? _brushWidth
          : (_mode == EditorMode.blur ? _blurWidth : _eraserWidth);

      setState(() {
        _currentStroke = EditorStroke(
          points: [imgPoint],
          color: _brushColor,
          strokeWidth: width,
          type: type,
        );
      });
    } else if (_mode == EditorMode.crop) {
      _handleCropPanStart(localPos, canvasSize);
    }
  }

  void _onPanUpdate(Offset localPos, Size canvasSize) {
    if (_currentStroke != null) {
      final imgPoint = _screenToImageOffset(localPos, canvasSize);
      if (imgPoint != null) {
        setState(() {
          _currentStroke!.points.add(imgPoint);
        });
      }
    } else if (_mode == EditorMode.crop) {
      _handleCropPanUpdate(localPos, canvasSize);
    }
  }

  void _onPanEnd() {
    if (_currentStroke != null) {
      setState(() {
        _strokes.add(_currentStroke!);
        _currentStroke = null;
      });
    }
    _cropDragHandle = null;
  }

  // Управление рамкой кадрирования
  String? _cropDragHandle; // 'tl', 'tr', 'bl', 'br', 'move'
  Offset? _cropStartPos;
  Rect? _cropStartRect;

  void _handleCropPanStart(Offset localPos, Size canvasSize) {
    if (_rawImage == null) return;
    final isRotated90 = _rotationQuarter % 2 != 0;
    final effW = isRotated90 ? _rawImage!.height.toDouble() : _rawImage!.width.toDouble();
    final effH = isRotated90 ? _rawImage!.width.toDouble() : _rawImage!.height.toDouble();

    final scale = math.min(canvasSize.width / effW, canvasSize.height / effH) * 0.94;
    final originX = (canvasSize.width - effW * scale) / 2;
    final originY = (canvasSize.height - effH * scale) / 2;

    final screenRect = Rect.fromLTWH(
      originX + _cropRect.left * effW * scale,
      originY + _cropRect.top * effH * scale,
      _cropRect.width * effW * scale,
      _cropRect.height * effH * scale,
    );

    const hit = 32.0;
    if ((localPos - screenRect.topLeft).distance < hit) {
      _cropDragHandle = 'tl';
    } else if ((localPos - screenRect.topRight).distance < hit) {
      _cropDragHandle = 'tr';
    } else if ((localPos - screenRect.bottomLeft).distance < hit) {
      _cropDragHandle = 'bl';
    } else if ((localPos - screenRect.bottomRight).distance < hit) {
      _cropDragHandle = 'br';
    } else if (screenRect.contains(localPos)) {
      _cropDragHandle = 'move';
    }

    _cropStartPos = localPos;
    _cropStartRect = _cropRect;
  }

  void _handleCropPanUpdate(Offset localPos, Size canvasSize) {
    if (_cropDragHandle == null || _cropStartPos == null || _cropStartRect == null || _rawImage == null) return;

    final isRotated90 = _rotationQuarter % 2 != 0;
    final effW = isRotated90 ? _rawImage!.height.toDouble() : _rawImage!.width.toDouble();
    final effH = isRotated90 ? _rawImage!.width.toDouble() : _rawImage!.height.toDouble();

    final scale = math.min(canvasSize.width / effW, canvasSize.height / effH) * 0.94;
    final dx = (localPos.dx - _cropStartPos!.dx) / (effW * scale);
    final dy = (localPos.dy - _cropStartPos!.dy) / (effH * scale);

    final orig = _cropStartRect!;
    double l = orig.left;
    double t = orig.top;
    double r = orig.right;
    double b = orig.bottom;

    if (_cropDragHandle == 'move') {
      final w = orig.width;
      final h = orig.height;
      l = (orig.left + dx).clamp(0.0, 1.0 - w);
      t = (orig.top + dy).clamp(0.0, 1.0 - h);
      r = l + w;
      b = t + h;
    } else if (_cropDragHandle == 'tl') {
      l = (orig.left + dx).clamp(0.0, orig.right - 0.1);
      t = (orig.top + dy).clamp(0.0, orig.bottom - 0.1);
    } else if (_cropDragHandle == 'tr') {
      r = (orig.right + dx).clamp(orig.left + 0.1, 1.0);
      t = (orig.top + dy).clamp(0.0, orig.bottom - 0.1);
    } else if (_cropDragHandle == 'bl') {
      l = (orig.left + dx).clamp(0.0, orig.right - 0.1);
      b = (orig.bottom + dy).clamp(orig.top + 0.1, 1.0);
    } else if (_cropDragHandle == 'br') {
      r = (orig.right + dx).clamp(orig.left + 0.1, 1.0);
      b = (orig.bottom + dy).clamp(orig.top + 0.1, 1.0);
    }

    setState(() {
      _cropRect = Rect.fromLTRB(l, t, r, b);
    });
  }

  // Панель параметров текущего режима
  Widget _buildModeParametersPanel() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      color: const Color(0xFF121216),
      child: AnimatedSwitcher(
        duration: const Duration(milliseconds: 200),
        child: _buildParameterContent(),
      ),
    );
  }

  Widget _buildParameterContent() {
    switch (_mode) {
      case EditorMode.crop:
        return Column(
          key: const ValueKey('crop_controls'),
          mainAxisSize: MainAxisSize.min,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceEvenly,
              children: [
                _buildPresetChip('Свободно', _aspectPreset == null, () => _setAspect(null)),
                _buildPresetChip('1:1', _aspectPreset == 1.0, () => _setAspect(1.0)),
                _buildPresetChip('4:3', _aspectPreset == 4.0 / 3.0, () => _setAspect(4.0 / 3.0)),
                _buildPresetChip('16:9', _aspectPreset == 16.0 / 9.0, () => _setAspect(16.0 / 9.0)),
              ],
            ),
            const SizedBox(height: 6),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                IconButton(
                  icon: const Icon(LucideIcons.rotate_ccw, color: Colors.white70),
                  tooltip: 'Повернуть влево (-90°)',
                  onPressed: _rotateCcw,
                ),
                IconButton(
                  icon: const Icon(LucideIcons.rotate_cw, color: Colors.white70),
                  tooltip: 'Повернуть вправо (+90°)',
                  onPressed: _rotateCw,
                ),
                const SizedBox(width: 8),
                IconButton(
                  icon: const Icon(LucideIcons.triangles_centerline_dashed_horizontal, color: Colors.white70),
                  tooltip: 'Отразить по горизонтали',
                  onPressed: _toggleFlipH,
                ),
                IconButton(
                  icon: const Icon(LucideIcons.triangles_centerline_dashed_vertical, color: Colors.white70),
                  tooltip: 'Отразить по вертикали',
                  onPressed: _toggleFlipV,
                ),
              ],
            ),
          ],
        );

      case EditorMode.brush:
        return Column(
          key: const ValueKey('brush_controls'),
          mainAxisSize: MainAxisSize.min,
          children: [
            Row(
              children: [
                const Text('Толщина: ', style: TextStyle(color: Colors.white70, fontSize: 12)),
                Expanded(
                  child: Slider(
                    value: _brushWidth,
                    min: 2,
                    max: 40,
                    activeColor: _brushColor,
                    onChanged: (val) => setState(() => _brushWidth = val),
                  ),
                ),
                Text('${_brushWidth.round()}px', style: const TextStyle(color: Colors.white70, fontSize: 12)),
              ],
            ),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: _brushColors.map((color) {
                final isSelected = _brushColor == color;
                return GestureDetector(
                  onTap: () => setState(() => _brushColor = color),
                  child: Container(
                    margin: const EdgeInsets.symmetric(horizontal: 6),
                    width: 28,
                    height: 28,
                    decoration: BoxDecoration(
                      color: color,
                      shape: BoxShape.circle,
                      border: Border.all(
                        color: isSelected ? const Color(0xFF6366F1) : Colors.white24,
                        width: isSelected ? 3.5 : 1.5,
                      ),
                      boxShadow: isSelected
                          ? [BoxShadow(color: color.withValues(alpha: 0.6), blurRadius: 8, spreadRadius: 1)]
                          : null,
                    ),
                  ),
                );
              }).toList(),
            ),
          ],
        );

      case EditorMode.blur:
        return Column(
          key: const ValueKey('blur_controls'),
          mainAxisSize: MainAxisSize.min,
          children: [
            Row(
              children: [
                const Text('Размер размытия: ', style: TextStyle(color: Colors.white70, fontSize: 12)),
                Expanded(
                  child: Slider(
                    value: _blurWidth,
                    min: 10,
                    max: 80,
                    activeColor: const Color(0xFF6366F1),
                    onChanged: (val) => setState(() => _blurWidth = val),
                  ),
                ),
                Text('${_blurWidth.round()}px', style: const TextStyle(color: Colors.white70, fontSize: 12)),
              ],
            ),
            const Text(
              '🕵️‍♂️ Замазывайте кистью номера, надписи, ценники или дефекты.',
              style: TextStyle(color: Colors.white54, fontSize: 11),
              textAlign: TextAlign.center,
            ),
          ],
        );

      case EditorMode.eraser:
        return Column(
          key: const ValueKey('eraser_controls'),
          mainAxisSize: MainAxisSize.min,
          children: [
            Row(
              children: [
                const Text('Размер ластика: ', style: TextStyle(color: Colors.white70, fontSize: 12)),
                Expanded(
                  child: Slider(
                    value: _eraserWidth,
                    min: 10,
                    max: 60,
                    activeColor: Colors.white,
                    onChanged: (val) => setState(() => _eraserWidth = val),
                  ),
                ),
                Text('${_eraserWidth.round()}px', style: const TextStyle(color: Colors.white70, fontSize: 12)),
              ],
            ),
            const Text(
              '🧯 Ластик восстанавливает исходную фотографию под проведенным пальцем.',
              style: TextStyle(color: Colors.white54, fontSize: 11),
              textAlign: TextAlign.center,
            ),
          ],
        );

      default:
        return const SizedBox.shrink();
    }
  }

  Widget _buildPresetChip(String label, bool isSelected, VoidCallback onTap) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
        decoration: BoxDecoration(
          color: isSelected ? const Color(0xFF6366F1) : const Color(0xFF27272A),
          borderRadius: BorderRadius.circular(12),
        ),
        child: Text(
          label,
          style: TextStyle(
            color: isSelected ? Colors.white : Colors.white70,
            fontSize: 12,
            fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
          ),
        ),
      ),
    );
  }

  // Нижний тулбар выбора режимов
  Widget _buildModeSelectorBar() {
    return Container(
      padding: const EdgeInsets.symmetric(vertical: 4),
      decoration: const BoxDecoration(
        color: Color(0xFF18181B),
        border: Border(top: BorderSide(color: Color(0xFF27272A))),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceAround,
        children: [
          _buildModeTab(EditorMode.crop, LucideIcons.crop, 'Кадрирование'),
          _buildModeTab(EditorMode.brush, LucideIcons.pencil, 'Маркер'),
          _buildModeTab(EditorMode.blur, LucideIcons.eye_off, 'Размытие'),
          _buildModeTab(EditorMode.eraser, LucideIcons.eraser, 'Ластик'),
        ],
      ),
    );
  }

  Widget _buildModeTab(EditorMode mode, IconData icon, String label) {
    final isSelected = _mode == mode;
    return InkWell(
      onTap: () => setState(() => _mode = mode),
      borderRadius: BorderRadius.circular(8),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              icon,
              color: isSelected ? const Color(0xFF6366F1) : Colors.white60,
              size: 22,
            ),
            const SizedBox(height: 2),
            Text(
              label,
              style: TextStyle(
                color: isSelected ? const Color(0xFF6366F1) : Colors.white60,
                fontSize: 10,
                fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
              ),
            ),
          ],
        ),
      ),
    );
  }

  // Нижние действия: Отмена / Скопировать / Сохранить
  Widget _buildBottomActionButtons() {
    return Container(
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 16),
      color: const Color(0xFF18181B),
      child: Row(
        children: [
          Expanded(
            child: OutlinedButton(
              onPressed: () => Navigator.of(context).pop(),
              style: OutlinedButton.styleFrom(
                foregroundColor: Colors.white70,
                side: const BorderSide(color: Color(0xFF3F3F46)),
                padding: const EdgeInsets.symmetric(vertical: 12),
              ),
              child: const Text('Отмена'),
            ),
          ),
          const SizedBox(width: 8),
          IconButton(
            tooltip: 'Скопировать результат в буфер без сохранения',
            style: IconButton.styleFrom(
              backgroundColor: const Color(0xFF27272A),
              foregroundColor: Colors.white,
            ),
            icon: const Icon(LucideIcons.copy, size: 20),
            onPressed: _rawImage != null && !_loading ? _handleCopy : null,
          ),
          const SizedBox(width: 8),
          Expanded(
            flex: 2,
            child: ElevatedButton.icon(
              onPressed: _rawImage != null && !_loading ? _handleSave : null,
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFF6366F1),
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(vertical: 12),
              ),
              icon: const Icon(LucideIcons.check, size: 18),
              label: const Text('Сохранить', style: TextStyle(fontWeight: FontWeight.bold)),
            ),
          ),
        ],
      ),
    );
  }
}

// CustomPainter отрисовки изображения, слоев размытия, маркера и рамки кропа
class _PhotoCanvasPainter extends CustomPainter {
  final ui.Image rawImage;
  final ui.Image? blurredImage;
  final int rotationQuarter;
  final bool flipH;
  final bool flipV;
  final Rect cropRect;
  final EditorMode mode;
  final List<EditorStroke> strokes;
  final EditorStroke? currentStroke;

  _PhotoCanvasPainter({
    required this.rawImage,
    required this.blurredImage,
    required this.rotationQuarter,
    required this.flipH,
    required this.flipV,
    required this.cropRect,
    required this.mode,
    required this.strokes,
    required this.currentStroke,
  });

  @override
  void paint(Canvas canvas, Size size) {
    final origW = rawImage.width.toDouble();
    final origH = rawImage.height.toDouble();

    final isRotated90 = rotationQuarter % 2 != 0;
    final effW = isRotated90 ? origH : origW;
    final effH = isRotated90 ? origW : origH;

    final scale = math.min(size.width / effW, size.height / effH) * 0.94;
    final originX = (size.width - effW * scale) / 2;
    final originY = (size.height - effH * scale) / 2;

    final imageScreenRect = Rect.fromLTWH(originX, originY, effW * scale, effH * scale);

    // Рисуем трансформированное изображение и штрихи
    canvas.save();
    canvas.translate(originX + effW * scale / 2, originY + effH * scale / 2);
    canvas.rotate(rotationQuarter * math.pi / 2);
    canvas.scale(flipH ? -scale : scale, flipV ? -scale : scale);
    canvas.translate(-origW / 2, -origH / 2);

    final rawRect = Rect.fromLTWH(0, 0, origW, origH);

    // 1. Рисуем сырую картинку
    canvas.drawImage(rawImage, Offset.zero, Paint());

    // 2. Слой штрихов
    final allStrokes = [...strokes, ?currentStroke];
    if (allStrokes.isNotEmpty && blurredImage != null) {
      canvas.saveLayer(rawRect, Paint());

      for (final stroke in allStrokes) {
        if (stroke.points.isEmpty) continue;

        if (stroke.type == StrokeType.blur) {
          canvas.saveLayer(rawRect, Paint());
          final maskPaint = Paint()
            ..style = PaintingStyle.stroke
            ..strokeCap = StrokeCap.round
            ..strokeJoin = StrokeJoin.round
            ..strokeWidth = stroke.strokeWidth;

          if (stroke.points.length == 1) {
            canvas.drawCircle(stroke.points.first, stroke.strokeWidth / 2, maskPaint..style = PaintingStyle.fill);
          } else {
            final path = Path()..moveTo(stroke.points.first.dx, stroke.points.first.dy);
            for (int i = 1; i < stroke.points.length; i++) {
              path.lineTo(stroke.points[i].dx, stroke.points[i].dy);
            }
            canvas.drawPath(path, maskPaint);
          }

          canvas.drawImage(blurredImage!, Offset.zero, Paint()..blendMode = BlendMode.srcIn);
          canvas.restore();
        } else if (stroke.type == StrokeType.brush) {
          final paint = Paint()
            ..color = stroke.color
            ..strokeWidth = stroke.strokeWidth
            ..style = PaintingStyle.stroke
            ..strokeCap = StrokeCap.round
            ..strokeJoin = StrokeJoin.round;

          if (stroke.points.length == 1) {
            canvas.drawCircle(stroke.points.first, stroke.strokeWidth / 2, paint..style = PaintingStyle.fill);
          } else {
            final path = Path()..moveTo(stroke.points.first.dx, stroke.points.first.dy);
            for (int i = 1; i < stroke.points.length; i++) {
              path.lineTo(stroke.points[i].dx, stroke.points[i].dy);
            }
            canvas.drawPath(path, paint);
          }
        } else if (stroke.type == StrokeType.eraser) {
          final erasePaint = Paint()
            ..blendMode = BlendMode.clear
            ..strokeWidth = stroke.strokeWidth
            ..style = PaintingStyle.stroke
            ..strokeCap = StrokeCap.round
            ..strokeJoin = StrokeJoin.round;

          if (stroke.points.length == 1) {
            canvas.drawCircle(stroke.points.first, stroke.strokeWidth / 2, erasePaint..style = PaintingStyle.fill);
          } else {
            final path = Path()..moveTo(stroke.points.first.dx, stroke.points.first.dy);
            for (int i = 1; i < stroke.points.length; i++) {
              path.lineTo(stroke.points[i].dx, stroke.points[i].dy);
            }
            canvas.drawPath(path, erasePaint);
          }
        }
      }

      canvas.restore();
    }

    canvas.restore();

    // 3. Если режим Crop — рисуем затемняющий оверлей и рамку кадрирования
    if (mode == EditorMode.crop) {
      final cropScreenRect = Rect.fromLTWH(
        imageScreenRect.left + cropRect.left * imageScreenRect.width,
        imageScreenRect.top + cropRect.top * imageScreenRect.height,
        cropRect.width * imageScreenRect.width,
        cropRect.height * imageScreenRect.height,
      );

      // Затемнение за пределами рамки
      final overlayPaint = Paint()..color = Colors.black.withValues(alpha: 0.65);
      final outsidePath = Path()
        ..addRect(Rect.fromLTWH(0, 0, size.width, size.height))
        ..addRect(cropScreenRect)
        ..fillType = PathFillType.evenOdd;
      canvas.drawPath(outsidePath, overlayPaint);

      // Рамка
      final borderPaint = Paint()
        ..color = Colors.white
        ..style = PaintingStyle.stroke
        ..strokeWidth = 2.0;
      canvas.drawRect(cropScreenRect, borderPaint);

      // Сетка третей
      final gridPaint = Paint()
        ..color = Colors.white.withValues(alpha: 0.35)
        ..style = PaintingStyle.stroke
        ..strokeWidth = 0.8;

      canvas.drawLine(
        Offset(cropScreenRect.left + cropScreenRect.width / 3, cropScreenRect.top),
        Offset(cropScreenRect.left + cropScreenRect.width / 3, cropScreenRect.bottom),
        gridPaint,
      );
      canvas.drawLine(
        Offset(cropScreenRect.left + 2 * cropScreenRect.width / 3, cropScreenRect.top),
        Offset(cropScreenRect.left + 2 * cropScreenRect.width / 3, cropScreenRect.bottom),
        gridPaint,
      );
      canvas.drawLine(
        Offset(cropScreenRect.left, cropScreenRect.top + cropScreenRect.height / 3),
        Offset(cropScreenRect.right, cropScreenRect.top + cropScreenRect.height / 3),
        gridPaint,
      );
      canvas.drawLine(
        Offset(cropScreenRect.left, cropScreenRect.top + 2 * cropScreenRect.height / 3),
        Offset(cropScreenRect.right, cropScreenRect.top + 2 * cropScreenRect.height / 3),
        gridPaint,
      );

      // Угловые ручки
      final handlePaint = Paint()
        ..color = const Color(0xFF6366F1)
        ..style = PaintingStyle.stroke
        ..strokeWidth = 3.5;
      const hLen = 16.0;

      // TL
      canvas.drawLine(cropScreenRect.topLeft, cropScreenRect.topLeft + const Offset(hLen, 0), handlePaint);
      canvas.drawLine(cropScreenRect.topLeft, cropScreenRect.topLeft + const Offset(0, hLen), handlePaint);
      // TR
      canvas.drawLine(cropScreenRect.topRight, cropScreenRect.topRight - const Offset(hLen, 0), handlePaint);
      canvas.drawLine(cropScreenRect.topRight, cropScreenRect.topRight + const Offset(0, hLen), handlePaint);
      // BL
      canvas.drawLine(cropScreenRect.bottomLeft, cropScreenRect.bottomLeft + const Offset(hLen, 0), handlePaint);
      canvas.drawLine(cropScreenRect.bottomLeft, cropScreenRect.bottomLeft - const Offset(0, hLen), handlePaint);
      // BR
      canvas.drawLine(cropScreenRect.bottomRight, cropScreenRect.bottomRight - const Offset(hLen, 0), handlePaint);
      canvas.drawLine(cropScreenRect.bottomRight, cropScreenRect.bottomRight - const Offset(0, hLen), handlePaint);
    }
  }

  @override
  bool shouldRepaint(covariant _PhotoCanvasPainter oldDelegate) => true;
}
