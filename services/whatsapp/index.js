const {
  default: makeWASocket,
  useMultiFileAuthState,
  DisconnectReason,
  fetchLatestBaileysVersion,
} = require('@whiskeysockets/baileys')
const express = require('express')
const qrcode = require('qrcode-terminal')
const pino = require('pino')

const PORT  = process.env.WHATSAPP_SERVICE_PORT  || 3210
const TOKEN = process.env.WHATSAPP_SERVICE_TOKEN || ''
const AUTH_DIR = process.env.WHATSAPP_AUTH_DIR   || './auth_info'

const logger = pino({ level: 'silent' })
const app = express()
app.use(express.json())

// Bearer-token auth (пропускаем /health без токена)
app.use((req, res, next) => {
  if (req.path === '/health') return next()
  if (!TOKEN) return next() // токен не задан — открытый режим
  const auth = req.headers['authorization'] || ''
  if (auth !== `Bearer ${TOKEN}`) {
    return res.status(401).json({ error: 'Unauthorized' })
  }
  next()
})

let sock = null
let isReady = false

// ── WhatsApp connection ───────────────────────────────────────────────────────

async function connectToWhatsApp() {
  const { state, saveCreds } = await useMultiFileAuthState(AUTH_DIR)
  const { version } = await fetchLatestBaileysVersion()

  sock = makeWASocket({
    version,
    logger,
    auth: state,
    printQRInTerminal: false,
  })

  sock.ev.on('creds.update', saveCreds)

  sock.ev.on('connection.update', ({ connection, lastDisconnect, qr }) => {
    if (qr) {
      console.log('\n📱 Отсканируйте QR-код в WhatsApp → Связанные устройства:\n')
      qrcode.generate(qr, { small: true })
    }

    if (connection === 'open') {
      isReady = true
      console.log('✅ WhatsApp подключён и готов к работе')
    }

    if (connection === 'close') {
      isReady = false
      const code = lastDisconnect?.error?.output?.statusCode
      const shouldReconnect = code !== DisconnectReason.loggedOut
      console.log(`⚠️  Соединение закрыто (code=${code}). Переподключение: ${shouldReconnect}`)
      if (shouldReconnect) {
        setTimeout(connectToWhatsApp, 3000)
      }
    }
  })
}

// ── HTTP API ──────────────────────────────────────────────────────────────────

// GET /health — проверка готовности
app.get('/health', (req, res) => {
  if (isReady) {
    res.json({ status: 'ok', whatsapp: 'connected' })
  } else {
    res.status(503).json({ status: 'error', whatsapp: 'not connected' })
  }
})

// POST /send — отправить сообщение
// Body: { "to": "+998901234567", "message": "Your OTP: 123456" }
app.post('/send', async (req, res) => {
  const { to, message } = req.body

  if (!to || !message) {
    return res.status(400).json({ error: 'to and message are required' })
  }

  if (!isReady || !sock) {
    return res.status(503).json({ error: 'WhatsApp not connected' })
  }

  try {
    // Форматировать номер в JID: +998901234567 -> 998901234567@s.whatsapp.net
    const jid = to.replace(/[^\d]/g, '') + '@s.whatsapp.net'
    await sock.sendMessage(jid, { text: message })
    console.log(`📤 Сообщение отправлено на ${to}`)
    res.json({ success: true })
  } catch (err) {
    console.error('Ошибка отправки:', err.message)
    res.status(500).json({ error: err.message })
  }
})

// ── Запуск ────────────────────────────────────────────────────────────────────

app.listen(PORT, () => {
  console.log(`🚀 WhatsApp микросервис запущен на порту ${PORT}`)
  connectToWhatsApp()
})

// Корректное завершение при SIGTERM/SIGINT (Go закрывает процесс)
process.on('SIGTERM', () => {
  console.log('WhatsApp сервис завершается...')
  process.exit(0)
})
process.on('SIGINT', () => {
  process.exit(0)
})
