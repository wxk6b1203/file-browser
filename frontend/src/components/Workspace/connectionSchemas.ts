export type ConnectionFieldKind = 'text' | 'password' | 'number' | 'textarea' | 'switch'

export interface ConnectionFieldSchema {
  key: string
  label: string
  kind: ConnectionFieldKind
  required?: boolean
  placeholder?: string
  min?: number
}

export const DRIVER_FIELD_SCHEMAS: Record<string, ConnectionFieldSchema[]> = {
  Local: [
    { key: 'rootPath', label: 'Root Path', kind: 'text', placeholder: 'C:/Users/wxk6b' },
  ],
  S3: [
    { key: 'region', label: 'Region', kind: 'text', required: true, placeholder: 'ap-southeast-1' },
    { key: 'bucket', label: 'Bucket', kind: 'text', required: true, placeholder: 'my-bucket' },
    { key: 'accessKeyID', label: 'Access Key ID', kind: 'text', required: true },
    { key: 'accessKeySecret', label: 'Access Key Secret', kind: 'password', required: true },
    { key: 'sessionToken', label: 'Session Token', kind: 'textarea' },
    { key: 'endpoint', label: 'Endpoint', kind: 'text', placeholder: 'https://s3.amazonaws.com' },
    { key: 'prefix', label: 'Prefix', kind: 'text', placeholder: 'path/to/root/' },
    { key: 'forcePathStyle', label: 'Force Path Style', kind: 'switch' },
    { key: 'disableSSL', label: 'Disable SSL', kind: 'switch' },
  ],
  OSS: [
    { key: 'region', label: 'Region', kind: 'text', required: true, placeholder: 'cn-hangzhou' },
    { key: 'bucket', label: 'Bucket', kind: 'text', required: true, placeholder: 'my-bucket' },
    { key: 'accessKeyID', label: 'Access Key ID', kind: 'text', required: true },
    { key: 'accessKeySecret', label: 'Access Key Secret', kind: 'password', required: true },
    { key: 'securityToken', label: 'Security Token', kind: 'textarea' },
    { key: 'endpoint', label: 'Endpoint', kind: 'text', placeholder: 'https://oss-cn-hangzhou.aliyuncs.com' },
    { key: 'prefix', label: 'Prefix', kind: 'text', placeholder: 'path/to/root/' },
    { key: 'forcePathStyle', label: 'Force Path Style', kind: 'switch' },
    { key: 'useCName', label: 'Use CName', kind: 'switch' },
    { key: 'disableSSL', label: 'Disable SSL', kind: 'switch' },
  ],
  SFTP: [
    { key: 'address', label: 'Address', kind: 'text', required: true, placeholder: '127.0.0.1' },
    { key: 'port', label: 'Port', kind: 'number', min: 1, placeholder: '22' },
    { key: 'username', label: 'Username', kind: 'text', required: true },
    { key: 'password', label: 'Password', kind: 'password' },
    { key: 'privateKeyPath', label: 'Private Key Path', kind: 'text', placeholder: '~/.ssh/id_rsa' },
    { key: 'privateKey', label: 'Private Key Text', kind: 'textarea', placeholder: '-----BEGIN OPENSSH PRIVATE KEY-----' },
    { key: 'passphrase', label: 'Passphrase', kind: 'password' },
    { key: 'rootPath', label: 'Root Path', kind: 'text', placeholder: '/home/user' },
    { key: 'dialTimeoutSec', label: 'Dial Timeout', kind: 'number', min: 1, placeholder: '30' },
  ],
}

export const DRIVER_DEFAULT_CONFIG: Record<string, Record<string, any>> = {
  Local: {
    rootPath: '',
  },
  S3: {
    region: '',
    bucket: '',
    accessKeyID: '',
    accessKeySecret: '',
    sessionToken: '',
    endpoint: '',
    prefix: '',
    forcePathStyle: false,
    disableSSL: false,
  },
  OSS: {
    region: '',
    bucket: '',
    accessKeyID: '',
    accessKeySecret: '',
    securityToken: '',
    endpoint: '',
    prefix: '',
    forcePathStyle: false,
    useCName: false,
    disableSSL: false,
  },
  SFTP: {
    address: '',
    port: 22,
    username: '',
    password: '',
    privateKey: '',
    privateKeyPath: '',
    passphrase: '',
    rootPath: '',
    dialTimeoutSec: 30,
  },
}

export function createDriverConfig(driver: string) {
  return { ...(DRIVER_DEFAULT_CONFIG[driver] ?? {}) }
}
