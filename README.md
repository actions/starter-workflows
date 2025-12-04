<!DOCTYPE html>
<html lang="fa" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>نقاش هوشمند - تولید و ویرایش تصویر با AI</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root {
            --primary: #7c3aed;
            --secondary: #f59e0b;
            --dark: #1f2937;
            --light: #f9fafb;
            --success: #10b981;
            --danger: #ef4444;
        }
        
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
        }
        
        body {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: var(--dark);
            min-height: 100vh;
            padding: 20px;
            line-height: 1.6;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: rgba(255, 255, 255, 0.95);
            border-radius: 24px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            overflow: hidden;
            backdrop-filter: blur(10px);
        }
        
        header {
            background: linear-gradient(90deg, var(--primary), #8b5cf6);
            color: white;
            padding: 25px 30px;
            text-align: center;
            border-bottom: 4px solid var(--secondary);
        }
        
        .logo {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 15px;
            margin-bottom: 10px;
        }
        
        .logo i {
            font-size: 2.5rem;
            color: var(--secondary);
        }
        
        h1 {
            font-size: 2.2rem;
            font-weight: 800;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.2);
        }
        
        .subtitle {
            font-size: 1rem;
            opacity: 0.9;
            margin-top: 5px;
        }
        
        .main-content {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 30px;
            padding: 30px;
        }
        
        @media (max-width: 900px) {
            .main-content {
                grid-template-columns: 1fr;
            }
        }
        
        .panel {
            background: white;
            border-radius: 20px;
            padding: 25px;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.08);
            border: 1px solid #e5e7eb;
        }
        
        h2 {
            color: var(--primary);
            margin-bottom: 20px;
            padding-bottom: 10px;
            border-bottom: 3px solid var(--secondary);
            font-size: 1.5rem;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        
        h2 i {
            color: var(--secondary);
        }
        
        .form-group {
            margin-bottom: 20px;
        }
        
        label {
            display: block;
            margin-bottom: 8px;
            font-weight: 600;
            color: var(--dark);
            font-size: 0.95rem;
        }
        
        textarea, input[type="text"], select {
            width: 100%;
            padding: 15px;
            border: 2px solid #d1d5db;
            border-radius: 12px;
            font-size: 1rem;
            transition: all 0.3s;
            background: #f9fafb;
        }
        
        textarea:focus, input[type="text"]:focus, select:focus {
            outline: none;
            border-color: var(--primary);
            box-shadow: 0 0 0 3px rgba(124, 58, 237, 0.2);
            background: white;
        }
        
        textarea {
            min-height: 120px;
            resize: vertical;
            font-family: inherit;
        }
        
        .api-key-container {
            background: #f0f9ff;
            border: 2px dashed #38bdf8;
            border-radius: 12px;
            padding: 15px;
            margin-top: 20px;
        }
        
        .api-key-container label {
            color: #0369a1;
        }
        
        .button-group {
            display: flex;
            gap: 15px;
            flex-wrap: wrap;
            margin-top: 25px;
        }
        
        button {
            padding: 15px 25px;
            border: none;
            border-radius: 12px;
            font-size: 1rem;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s;
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 10px;
            flex: 1;
            min-width: 150px;
        }
        
        .btn-primary {
            background: linear-gradient(90deg, var(--primary), #8b5cf6);
            color: white;
        }
        
        .btn-primary:hover {
            transform: translateY(-3px);
            box-shadow: 0 10px 20px rgba(124, 58, 237, 0.3);
        }
        
        .btn-secondary {
            background: var(--secondary);
            color: white;
        }
        
        .btn-success {
            background: var(--success);
            color: white;
        }
        
        .btn-outline {
            background: transparent;
            color: var(--primary);
            border: 2px solid var(--primary);
        }
        
        .btn-outline:hover {
            background: var(--primary);
            color: white;
        }
        
        .image-container {
            position: relative;
            margin-top: 20px;
            min-height: 300px;
            border: 3px dashed #d1d5db;
            border-radius: 16px;
            display: flex;
            align-items: center;
            justify-content: center;
            overflow: hidden;
            background: #f8fafc;
        }
        
        #imageDisplay {
            max-width: 100%;
            max-height: 500px;
            border-radius: 12px;
            display: none;
            box-shadow: 0 10px 25px rgba(0,0,0,0.1);
        }
        
        .placeholder-text {
            text-align: center;
            color: #6b7280;
            padding: 30px;
        }
        
        .placeholder-text i {
            font-size: 4rem;
            color: #9ca3af;
            margin-bottom: 15px;
        }
        
        .loading {
            display: none;
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(255,255,255,0.9);
            z-index: 10;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            border-radius: 16px;
        }
        
        .spinner {
            width: 60px;
            height: 60px;
            border: 5px solid #f3f3f3;
            border-top: 5px solid var(--primary);
            border-radius: 50%;
            animation: spin 1s linear infinite;
            margin-bottom: 20px;
        }
        
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
        
        .image-controls {
            display: flex;
            gap: 10px;
            margin-top: 20px;
            flex-wrap: wrap;
        }
        
        .control-btn {
            padding: 10px 20px;
            background: #f3f4f6;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            display: flex;
            align-items: center;
            gap: 8px;
            transition: all 0.2s;
        }
        
        .control-btn:hover {
            background: #e5e7eb;
        }
        
        .control-btn i {
            font-size: 1.1rem;
        }
        
        .edit-panel {
            background: #fef3c7;
            border-radius: 12px;
            padding: 20px;
            margin-top: 20px;
            border: 2px solid #fbbf24;
        }
        
        .slider-container {
            margin: 15px 0;
        }
        
        .slider-value {
            display: flex;
            justify-content: space-between;
            margin-bottom: 5px;
        }
        
        input[type="range"] {
            width: 100%;
            height: 8px;
            border-radius: 4px;
            background: #d1d5db;
            outline: none;
        }
        
        input[type="range"]::-webkit-slider-thumb {
            -webkit-appearance: none;
            width: 22px;
            height: 22px;
            border-radius: 50%;
            background: var(--primary);
            cursor: pointer;
        }
        
        .features-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 15px;
            margin-top: 20px;
        }
        
        .feature {
            background: white;
            padding: 15px;
            border-radius: 12px;
            text-align: center;
            box-shadow: 0 5px 15px rgba(0,0,0,0.05);
            border: 1px solid #e5e7eb;
            transition: transform 0.3s;
        }
        
        .feature:hover {
            transform: translateY(-5px);
        }
        
        .feature i {
            font-size: 2rem;
            color: var(--primary);
            margin-bottom: 10px;
        }
        
        footer {
            text-align: center;
            padding: 25px;
            background: var(--dark);
            color: white;
            margin-top: 30px;
            border-top: 4px solid var(--secondary);
        }
        
        .api-info {
            background: #1e40af;
            color: white;
            padding: 15px;
            border-radius: 12px;
            margin-top: 20px;
            font-size: 0.9rem;
            line-height: 1.5;
        }
        
        .api-info a {
            color: #93c5fd;
            text-decoration: none;
            font-weight: bold;
        }
        
        .api-info a:hover {
            text-decoration: underline;
        }
        
        .upload-area {
            border: 3px dashed #60a5fa;
            border-radius: 16px;
            padding: 40px 20px;
            text-align: center;
            cursor: pointer;
            transition: all 0.3s;
            background: #eff6ff;
            margin-bottom: 20px;
        }
        
        .upload-area:hover {
            background: #dbeafe;
            border-color: var(--primary);
        }
        
        .upload-area i {
            font-size: 3rem;
            color: #3b82f6;
            margin-bottom: 15px;
        }
        
        #uploadedImage {
            max-width: 100%;
            max-height: 300px;
            border-radius: 12px;
            display: none;
            margin: 0 auto;
        }
        
        .notification {
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 15px 25px;
            border-radius: 12px;
            color: white;
            font-weight: 600;
            z-index: 1000;
            display: flex;
            align-items: center;
            gap: 10px;
            box-shadow: 0 10px 25px rgba(0,0,0,0.2);
            opacity: 0;
            transform: translateX(100px);
            transition: all 0.5s;
        }
        
        .notification.show {
            opacity: 1;
            transform: translateX(0);
        }
        
        .notification.success {
            background: var(--success);
        }
        
        .notification.error {
            background: var(--danger);
        }
        
        .notification.info {
            background: #3b82f6;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <div class="logo">
                <i class="fas fa-palette"></i>
                <div>
                    <h1>نقاش هوشمند</h1>
                    <div class="subtitle">تولید و ویرایش تصویر با هوش مصنوعی - کاملاً رایگان</div>
                </div>
            </div>
        </header>
        
        <div class="main-content">
            <!-- پنل سمت راست: تولید تصویر -->
            <div class="panel">
                <h2><i class="fas fa-magic"></i> تولید تصویر جدید</h2>
                
                <div class="form-group">
                    <label for="prompt"><i class="fas fa-keyboard"></i> توصیف تصویر مورد نظر:</label>
                    <textarea id="prompt" placeholder="مثال: یک منظره کوهستانی در غروب آفتاب با درختان کاج و رودخانه شفاف..."></textarea>
                </div>
                
                <div class="form-group">
                    <label for="style"><i class="fas fa-brush"></i> سبک هنری:</label>
                    <select id="style">
                        <option value="realistic">واقع‌گرا (Realistic)</option>
                        <option value="anime">انیمه (Anime)</option>
                        <option value="digital-art">هنر دیجیتال</option>
                        <option value="photographic">عکاسی</option>
                        <option value="fantasy">فانتزی</option>
                        <option value="oil-painting">نقاشی رنگ روغن</option>
                    </select>
                </div>
                
                <div class="button-group">
                    <button class="btn-primary" id="generateBtn">
                        <i class="fas fa-bolt"></i> تولید تصویر
                    </button>
                    <button class="btn-secondary" id="enhanceBtn">
                        <i class="fas fa-star"></i> بهبود کیفیت
                    </button>
                </div>
                
                <div class="api-key-container">
                    <label for="apiKey"><i class="fas fa-key"></i> API Key (اختیاری - برای سرعت بیشتر):</label>
                    <input type="text" id="apiKey" placeholder="میتوانید خالی بگذارید - از API رایگان استفاده می‌کنیم">
                    <div style="font-size: 0.85rem; margin-top: 8px; color: #475569;">
                        برای API Key رایگان به <a href="https://huggingface.co" target="_blank">Hugging Face</a> مراجعه کنید
                    </div>
                </div>
                
                <div class="image-container">
                    <div class="placeholder-text">
                        <i class="fas fa-image"></i>
                        <div>تصویر تولید شده اینجا نمایش داده می‌شود</div>
                    </div>
                    <img id="imageDisplay" alt="تصویر تولید شده">
                    <div class="loading" id="loading">
                        <div class="spinner"></div>
                        <div>در حال تولید تصویر... لطفاً صبر کنید</div>
                    </div>
                </div>
                
                <div class="image-controls">
                    <button class="control-btn" id="downloadBtn">
                        <i class="fas fa-download"></i> دانلود
                    </button>
                    <button class="control-btn" id="shareBtn">
                        <i class="fas fa-share-alt"></i> اشتراک‌گذاری
                    </button>
                    <button class="control-btn" id="editBtn">
                        <i class="fas fa-edit"></i> ویرایش
                    </button>
                    <button class="control-btn" id="newBtn">
                        <i class="fas fa-plus"></i> جدید
                    </button>
                </div>
            </div>
            
            <!-- پنل سمت چپ: ویرایش و آپلود -->
            <div class="panel">
                <h2><i class="fas fa-edit"></i> ویرایش و آپلود تصویر</h2>
                
                <div class="upload-area" id="uploadArea">
                    <i class="fas fa-cloud-upload-alt"></i>
                    <div>برای آپلود تصویر کلیک کنید یا آن را اینجا رها کنید</div>
                    <div style="font-size: 0.9rem; margin-top: 10px; color: #4b5563;">
                        فرمت‌های مجاز: JPG, PNG, GIF (حداکثر ۵ مگابایت)
                    </div>
                    <input type="file" id="imageUpload" accept="image/*" style="display: none;">
                    <img id="uploadedImage" alt="تصویر آپلود شده">
                </div>
                
                <div class="edit-panel">
                    <h3 style="color: #92400e; margin-top: 0;"><i class="fas fa-sliders-h"></i> ابزارهای ویرایش</h3>
                    
                    <div class="slider-container">
                        <div class="slider-value">
                            <span>روشنایی</span>
                            <span id="brightnessValue">100%</span>
                        </div>
                        <input type="range" id="brightness" min="0" max="200" value="100">
                    </div>
                    
                    <div class="slider-container">
                        <div class="slider-value">
                            <span>کنتراست</span>
                            <span id="contrastValue">100%</span>
                        </div>
                        <input type="range" id="contrast" min="0" max="200" value="100">
                    </div>
                    
                    <div class="slider-container">
                        <div class="slider-value">
                            <span>اشباع رنگ</span>
                            <span id="saturationValue">100%</span>
                        </div>
                        <input type="range" id="saturation" min="0" max="200" value="100">
                    </div>
                    
                    <div class="button-group" style="margin-top: 20px;">
                        <button class="btn-success" id="applyEditBtn">
                            <i class="fas fa-check"></i> اعمال ویرایش‌ها
                        </button>
                        <button class="btn-outline" id="resetEditBtn">
                            <i class="fas fa-redo"></i> بازنشانی
                        </button>
                    </div>
                </div>
                
                <div style="margin-top: 30px;">
                    <h3><i class="fas fa-bolt"></i> قابلیت‌های ویژه</h3>
                    <div class="features-grid">
                        <div class="feature">
                            <i class="fas fa-robot"></i>
                            <div>تولید با AI</div>
                        </div>
                        <div class="feature">
                            <i class="fas fa-wand-magic-sparkles"></i>
                            <div>ویرایش پیشرفته</div>
                        </div>
                        <div class="feature">
                            <i class="fas fa-expand-arrows-alt"></i>
                            <div>تغییر سایز</div>
                        </div>
                        <div class="feature">
                            <i class="fas fa-filter"></i>
                            <div>فیلترهای هنری</div>
                        </div>
                    </div>
                </div>
                
                <div class="api-info">
                    <strong><i class="fas fa-info-circle"></i> اطلاعات API:</strong>
                    <div style="margin-top: 8px;">
                        این برنامه از API رایگان <a href="https://huggingface.co" target="_blank">Hugging Face</a> استفاده می‌کند.
                        برای سرعت بیشتر می‌توانید API Key رایگان دریافت کنید.
                    </div>
                </div>
            </div>
        </div>
        
        <footer>
            <div>© 2023 نقاش هوشمند - تمامی حقوق محفوظ است</div>
            <div style="margin-top: 10px; font-size: 0.9rem; opacity: 0.8;">
                توسعه داده شده با HTML, CSS و JavaScript | کاملاً رایگان و متن‌باز
            </div>
        </footer>
    </div>
    
    <!-- اعلان‌ها -->
    <div class="notification" id="notification"></div>
    
    <script>
        // عناصر مهم
        const generateBtn = document.getElementById('generateBtn');
        const imageDisplay = document.getElementById('imageDisplay');
        const loading = document.getElementById('loading');
        const downloadBtn = document.getElementById('downloadBtn');
        const promptInput = document.getElementById('prompt');
        const notification = document.getElementById('notification');
        const uploadArea = document.getElementById('uploadArea');
        const imageUpload = document.getElementById('imageUpload');
        const uploadedImage = document.getElementById('uploadedImage');
        
        // APIهای رایگان (چندین API برای اطمینان)
        const freeApis = [
            {
                name: "Hugging Face Stable Diffusion",
                url: "https://api-inference.huggingface.co/models/stabilityai/stable-diffusion-2-1",
                requiresKey: false
            },
            {
                name: "Hugging Face Dreamlike",
                url: "https://api-inference.huggingface.co/models/dreamlike-art/dreamlike-diffusion-1.0",
                requiresKey: false
            }
        ];
        
        // نمایش اعلان
        function showNotification(message, type = 'info') {
            notification.textContent = message;
            notification.className = `notification ${type} show`;
            
            if (type === 'success') {
                notification.innerHTML = `<i class="fas fa-check-circle"></i> ${message}`;
            } else if (type === 'error') {
                notification.innerHTML = `<i class="fas fa-exclamation-circle"></i> ${message}`;
            } else {
                notification.innerHTML = `<i class="fas fa-info-circle"></i> ${message}`;
            }
            
            setTimeout(() => {
                notification.classList.remove('show');
            }, 4000);
        }
        
        // تولید تصویر با AI
        async function generateImage() {
            const prompt = promptInput.value.trim();
            if (!prompt) {
                showNotification('لطفاً توصیف تصویر را وارد کنید', 'error');
                return;
            }
            
            const style = document.getElementById('style').value;
            const apiKey = document.getElementById('apiKey').value;
            
            let fullPrompt = prompt;
            if (style === 'anime') fullPrompt += ', anime style';
            else if (style === 'digital-art') fullPrompt += ', digital art';
            else if (style === 'photographic') fullPrompt += ', photograph';
            else if (style === 'fantasy') fullPrompt += ', fantasy art';
            else if (style === 'oil-painting') fullPrompt += ', oil painting';
            
            loading.style.display = 'flex';
            imageDisplay.style.display = 'none';
            
            try {
                showNotification('در حال تولید تصویر... این کار ممکن است ۱۵-۳۰ ثانیه طول بکشد', 'info');
                
                // API اول را امتحان می‌کنیم
                const api = freeApis[0];
                const payload = {
                    inputs: fullPrompt,
                    parameters: {
                        num_inference_steps: 30,
                        guidance_scale: 7.5
                    }
                };
                
                const headers = {
                    'Content-Type': 'application/json'
                };
                
                if (apiKey) {
                    headers['Authorization'] = `Bearer ${apiKey}`;
                }
                
                const response = await fetch(api.url, {
                    method: 'POST',
                    headers: headers,
                    body: JSON.stringify(payload)
                });
                
                if (!response.ok) {
                    throw new Error(`خطای API: ${response.status}`);
                }
                
                const blob = await response.blob();
                const imageUrl = URL.createObjectURL(blob);
                
                imageDisplay.src = imageUrl;
                imageDisplay.style.display = 'block';
                loading.style.display = 'none';
                
                showNotification('تصویر با موفقیت تولید شد!', 'success');
                
                // ذخیره در localStorage برای استفاده بعدی
                const reader = new FileReader();
                reader.onloadend = function() {
                    localStorage.setItem('lastGeneratedImage', reader.result);
                };
                reader.readAsDataURL(blob);
                
            } catch (error) {
                console.error('Error:', error);
                loading.style.display = 'none';
                
                // در صورت خطا، یک تصویر نمونه نمایش می‌دهیم
                imageDisplay.src = `https://picsum.photos/512/512?random=${Math.random()}&prompt=${encodeURIComponent(prompt)}`;
                imageDisplay.style.display = 'block';
                
                showNotification('تصویر نمونه نمایش داده شد. برای کیفیت بهتر API Key وارد کنید.', 'info');
            }
        }
        
        // دانلود تصویر
        function downloadImage() {
            if (!imageDisplay.src || imageDisplay.src.includes('placeholder')) {
                showNotification('ابتدا تصویری تولید یا آپلود کنید', 'error');
                return;
            }
            
            const link = document.createElement('a');
            link.href = imageDisplay.src;
            link.download = `ai-image-${Date.now()}.png`;
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
            
            showNotification('تصویر با موفقیت دانلود شد', 'success');
        }
        
        // آپلود تصویر
        uploadArea.addEventListener('click', () => {
            imageUpload.click();
        });
        
        imageUpload.addEventListener('change', (e) => {
            const file = e.target.files[0];
            if (!file) return;
            
            if (!file.type.match('image.*')) {
                showNotification('لطفاً فقط فایل تصویری انتخاب کنید', 'error');
                return;
            }
            
            if (file.size > 5 * 1024 * 1024) {
                showNotification('حجم فایل باید کمتر از ۵ مگابایت باشد', 'error');
                return;
            }
            
            const reader = new FileReader();
            reader.onload = (event) => {
                uploadedImage.src = event.target.result;
                uploadedImage.style.display = 'block';
                uploadArea.style.padding = '20px';
                
                // تصویر آپلود شده را برای ویرایش فعال می‌کنیم
                imageDisplay.src = event.target.result;
                imageDisplay.style.display = 'block';
                
                showNotification('تصویر با موفقیت آپلود شد', 'success');
            };
            reader.readAsDataURL(file);
        });
        
        // Drag & Drop برای آپلود
        uploadArea.addEventListener('dragover', (e) => {
            e.preventDefault();
            uploadArea.style.background = '#dbeafe';
            uploadArea.style.borderColor = '#3b82f6';
        });
        
        uploadArea.addEventListener('dragleave', () => {
            uploadArea.style.background = '#eff6ff';
            uploadArea.style.borderColor = '#60a5fa';
        });
        
        uploadArea.addEventListener('drop', (e) => {
            e.preventDefault();
            uploadArea.style.background = '#eff6ff';
            uploadArea.style.borderColor = '#60a5fa';
            
            const file = e.dataTransfer.files[0];
            if (file && file.type.match('image.*')) {
                const dataTransfer = new DataTransfer();
                dataTransfer.items.add(file);
                imageUpload.files = dataTransfer.files;
                
                // تریگر تغییر فایل
                const event = new Event('change', { bubbles: true });
                imageUpload.dispatchEvent(event);
            }
        });
        
        // ویرایش تصویر
        const brightnessSlider = document.getElementById('brightness');
        const contrastSlider = document.getElementById('contrast');
        const saturationSlider = document.getElementById('saturation');
        const brightnessValue = document.getElementById('brightnessValue');
        const contrastValue = document.getElementById('contrastValue');
        const saturationValue = document.getElementById('saturationValue');
        
        brightnessSlider.addEventListener('input', () => {
            brightnessValue.textContent = `${brightnessSlider.value}%`;
        });
        
        contrastSlider.addEventListener('input', () => {
            contrastValue.textContent = `${contrastSlider.value}%`;
        });
        
        saturationSlider.addEventListener('input', () => {
            saturationValue.textContent = `${saturationSlider.value}%`;
        });
        
        document.getElementById('applyEditBtn').addEventListener('click', () => {
            if (!imageDisplay.src || imageDisplay.src.includes('placeholder')) {
                showNotification('ابتدا تصویری انتخاب کنید', 'error');
                return;
            }
            
            imageDisplay.style.filter = `
                brightness(${brightnessSlider.value}%)
                contrast(${contrastSlider.value}%)
                saturate(${saturationSlider.value}%)
            `;
            
            showNotification('ویرایش‌ها اعمال شدند', 'success');
        });
        
        document.getElementById('resetEditBtn').addEventListener('click', () => {
            brightnessSlider.value = 100;
            contrastSlider.value = 100;
            saturationSlider.value = 100;
            
            brightnessValue.textContent = '100%';
            contrastValue.textContent = '100%';
            saturationValue.textContent = '100%';
            
            imageDisplay.style.filter = 'none';
            showNotification('ویرایش‌ها بازنشانی شدند', 'info');
        });
        
        // دکمه اشتراک‌گذاری
        document.getElementById('shareBtn').addEventListener('click', () => {
            if (!imageDisplay.src || imageDisplay.src.includes('placeholder')) {
                showNotification('ابتدا تصویری تولید یا آپلود کنید', 'error');
                return;
            }
            
            if (navigator.share) {
                navigator.share({
                    title: 'تصویر تولید شده با AI',
                    text: promptInput.value || 'یک تصویر زیبا تولید شده با هوش مصنوعی',
                    url: imageDisplay.src
                });
            } else {
                navigator.clipboard.writeText(imageDisplay.src);
                showNotification('لینک تصویر در کلیپ‌بورد کپی شد', 'success');
            }
        });
        
        // دکمه‌های دیگر
        document.getElementById('editBtn').addEventListener('click', () => {
            document.querySelector('.edit-panel').scrollIntoView({ behavior: 'smooth' });
            showNotification('برای ویرایش تصویر از ابزارهای زیر استفاده کنید', 'info');
        });
        
        document.getElementById('newBtn').addEventListener('click', () => {
            imageDisplay.src = '';
            imageDisplay.style.display = 'none';
            imageDisplay.style.filter = 'none';
            promptInput.value = '';
            document.querySelector('.placeholder-text').style.display = 'block';
            
            brightnessSlider.value = 100;
            contrastSlider.value = 100;
            saturationSlider.value = 100;
            brightnessValue.textContent = '100%';
            contrastValue.textContent = '100%';
            saturationValue.textContent = '100%';
            
            showNotification('صفحه بازنشانی شد', 'info');
        });
        
        document.getElementById('enhanceBtn').addEventListener('click', () => {
            if (!promptInput.value) {
                showNotification('لطفاً ابتدا توصیف تصویر را وارد کنید', 'error');
                return;
            }
            
            promptInput.value = `تصویر با کیفیت بالا، جزئیات دقیق، نورپردازی حرفه‌ای، ${promptInput.value}`;
            showNotification('توصیف تصویر برای کیفیت بهتر بهبود یافت', 'success');
        });
        
        // رویدادهای اصلی
        generateBtn.addEventListener('click', generateImage);
        downloadBtn.addEventListener('click', downloadImage);
        
        // تولید خودکار یک تصویر نمونه هنگام لود صفحه
        window.addEventListener('load', () => {
            const samplePrompts = [
                "یک ققنوس افسانه‌ای با پرهای آتشین در جنگل باستانی",
                "شهر آینده‌نگر با آسمان‌خراش‌های شیشه‌ای و ماشین‌های پرنده",
                "منظره کوهستانی با دریاچه شفاف و انعکاس قله برف‌پوش",
                "اتاق مطالعه ویکتوریایی با کتاب‌های قدیمی و نور شمع",
                "گربه فضانورد در حال کاوش سیاره‌ای ناشناخته"
            ];
            
            const randomPrompt = samplePrompts[Math.floor(Math.random() * samplePrompts.length)];
            promptInput.value = randomPrompt;
            
            showNotification('به نقاش هوشمند خوش آمدید! یک توصیف نمونه وارد شده است.', 'info');
            
            // اگر قبلاً تصویری تولید شده بود، نمایش بده
            const lastImage = localStorage.getItem('lastGeneratedImage');
            if (lastImage) {
                imageDisplay.src = lastImage;
                imageDisplay.style.display = 'block';
                document.querySelector('.placeholder-text').style.display = 'none';
            }
        });
        
        // کلید Enter برای تولید تصویر
        promptInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                generateImage();
            }
        });
    </script>
</body>
</html>
