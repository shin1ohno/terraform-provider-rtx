# 第21章: SNMP の設定

> 元PDFページ: 378-396

---

第 21 章
SNMP の設定
SNMP (Simple Network Management Protocol) の設定を行うことにより、 SNMP 管理ソフトウェアに対してネットワー
ク管理情報のモニタと変更を行うことができるようになります。このとき  ヤマハルーター  は SNMP エージェント
となります。
ヤマハルーター  は SNMPv1 、SNMPv2c 、SNMPv3 による通信に対応しています。また  MIB (Management information
Base) として  RFC1213 (MIB-II) とプライベート  MIB に対応しています。プライベート  MIB については以下の  URL
から参照することができます。
• YAMAHA private MIB: https://www.rtpro.yamaha.co.jp/RT/docs/mib/
SNMPv1 および  SNMPv2c では、コミュニティと呼ばれるグループの名前を相手に通知し、同じコミュニティに属す
るホスト間でのみ通信します。このとき、読み出し専用  (read-only) と読み書き可能  (read-write) の 2 つのアクセスモ
ードに対して別々にコミュニティ名を設定することができます。
このようにコミュニティ名はある種のパスワードとして機能しますが、その反面、コミュニティ名は必ず平文でネ
ットワーク上を流れるという特性があり、セキュリティ面では脆弱と言えます。よりセキュアな通信が必要な場合
は SNMPv3 の利用を推奨します。
SNMPv3 では通信内容の認証、および暗号化に対応しています。 SNMPv3 はコミュニティの概念を廃し、新たに
USM (User-based Security Model) と呼ばれるセキュリティモデルを利用することで、より高度なセキュリティを確保
しています。
ヤマハルーター  の状態を通知する  SNMP メッセージをトラップと呼びます。ヤマハルーター  では  SNMP 標準トラ
ップの他にも、一部機能で特定のイベントを通知するため独自のトラップを送信することがあります。なお、これ
らの独自トラップはプライベート  MIB として定義されています。
トラップの送信先ホストについては、各  SNMP バージョン毎に複数のホストを設定することができます。
SNMPv1 および  SNMPv2c で利用する読み出し専用と送信トラップ用のコミュニティ名は、共に初期値が  "public" と
なっています。 SNMP 管理ソフトウェア側も  "public" がコミュニティ名である場合が多いため、 当該バージョンの通
信でセキュリティを考慮する場合は適切なコミュニティ名に変更してください。ただし、上述の通りコミュニティ
名はネットワーク上を平文で流れますので、コミュニティ名にログインパスワードや管理パスワードを決して使用
しないよう注意してください。
工場出荷状態では、各  SNMP バージョンにおいてアクセスが一切できない状態となっています。また、トラップの
送信先ホストは設定されておらず、どこにもトラップを送信しません。
21.1 SNMPv1 によるアクセスを許可するホストの設定
[書式 ]
snmp  host host [ro_community  [rw_community ]]
no snmp  host [host]
[設定値及び初期値 ]
•host : SNMPv1 によるアクセスを許可するホスト
• [設定値 ] :
設定値 説明
ip_address 1個の IPアドレスまたは間にハイフン (-)をはさんだ
IPアドレス (範囲指定 )
lanN LANインターフェース名
bridgeN ブリッジインターフェース名
any すべてのホストからのアクセスを許可する
none すべてのホストからのアクセスを禁止する
• [初期値 ] : none
•ro_community
• [設定値 ] : 読み出し専用のコミュニティ名  (16 文字以内  )
• [初期値 ] : -
•rw_community378 | コマンドリファレンス  | SNMP の設定

• [設定値 ] : 読み書き可能なコミュニティ名  (16 文字以内  )
• [初期値 ] : -
[説明 ]
SNMPv1 によるアクセスを許可するホストを設定する。
'any' を設定した場合は任意のホストからの  SNMPv1 によるアクセスを許可する。
IPアドレスや lanN、bridgeNでホストを指定した場合には、同時にコミュニティ名も設定できる。 rw_community  パ
ラメータを省略した場合には、アクセスモードが読み書き可能であるアクセスが禁止される。 ro_community  パラメ
ータも省略した場合には、 snmp community read-only  コマンド、および  snmp community read-write  コマンドの設定
値が用いられる。
[ノート ]
host パラメーターにおける  IP アドレスの範囲指定と  LAN インターフェース名の指定は、 RTX5000 / RTX3500
Rev.14.00.17 以前では不可能。ブリッジインターフェースの指定は、 vRX Amazon EC2 版、 および  RTX5000 / RTX3500
Rev.14.00.17 以前では不可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.2 SNMPv1 の読み出し専用のコミュニティ名の設定
[書式 ]
snmp  community  read-only  name
no snmp  community  read-only
[設定値及び初期値 ]
•name
• [設定値 ] : コミュニティ名  (16 文字以内  )
• [初期値 ] : public
[説明 ]
SNMPv1 によるアクセスモードが読み出し専用であるコミュニティ名を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.3 SNMPv1 の読み書き可能なコミュニティ名の設定
[書式 ]
snmp  community  read-write  name
no snmp  community  read-write
[設定値及び初期値 ]
•name
• [設定値 ] : コミュニティ名  (16 文字以内  )
• [初期値 ] : -
[説明 ]
SNMPv1 によるアクセスモードが読み書き可能であるコミュニティ名を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.4 SNMPv1 トラップの送信先の設定
[書式 ]
snmp  trap host host [community ]
no snmp  trap host host
[設定値及び初期値 ]
•host
• [設定値 ] : SNMPv1 トラップの送信先ホストの  IP アドレス  (IPv4/IPv6)コマンドリファレンス  | SNMP の設定  | 379

• [初期値 ] : -
•community
• [設定値 ] : コミュニティ名  (16 文字以内  )
• [初期値 ] : -
[説明 ]
SNMPv1 トラップを送信するホストを指定する。コマンドを複数設定することで、複数のホストを同時に指定でき
る。トラップ送信時のコミュニティ名にはこのコマンドの  community  パラメータが用いられるが、省略されている
場合には  snmp trap community  コマンドの設定値が用いられる。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.5 SNMPv1 トラップのコミュニティ名の設定
[書式 ]
snmp  trap community  name
no snmp  trap community
[設定値及び初期値 ]
•name
• [設定値 ] : コミュニティ名  (16 文字以内  )
• [初期値 ] : public
[説明 ]
SNMPv1 トラップを送信する際のコミュニティ名を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.6 SNMPv2c によるアクセスを許可するホストの設定
[書式 ]
snmpv2c  host host [ro_community  [rw_community ]]
no snmpv2c  host [host]
[設定値及び初期値 ]
•host : SNMPv2c によるアクセスを許可するホスト
• [設定値 ] :
設定値 説明
ip_address 1個の IPアドレスまたは間にハイフン (-)をはさんだ
IPアドレス (範囲指定 )
lanN LANインターフェース名
bridgeN ブリッジインターフェース名
any すべてのホストからのアクセスを許可する
none すべてのホストからのアクセスを禁止する
• [初期値 ] : none
•ro_community
• [設定値 ] : 読み出し専用のコミュニティ名  (16 文字以内  )
• [初期値 ] : -
•rw_community
• [設定値 ] : 読み書き可能なコミュニティ名  (16 文字以内  )
• [初期値 ] : -
[説明 ]
SNMPv2c によるアクセスを許可するホストを設定する。
'any' を設定した場合は任意のホストからの  SNMPv2c によるアクセスを許可する。380 | コマンドリファレンス  | SNMP の設定

IPアドレスや lanN、bridgeNでホストを指定した場合には、同時にコミュニティ名も設定できる。 rw_community  パ
ラメータを省略した場合には、アクセスモードが読み書き可能であるアクセスが禁止される。 ro_community  パラメ
ータも省略した場合には、 snmpv2c community read-only  コマンド、および  snmpv2c community read-write  コマンド
の設定値が用いられる。
[ノート ]
host パラメーターにおける  IP アドレスの範囲指定と  LAN インターフェース名の指定は、 RTX5000 / RTX3500
Rev.14.00.17 以前では不可能。ブリッジインターフェースの指定は、 vRX Amazon EC2 版、 および  RTX5000 / RTX3500
Rev.14.00.17 以前では不可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.7 SNMPv2c の読み出し専用のコミュニティ名の設定
[書式 ]
snmpv2c  community  read-only  name
no snmpv2c  community  read-only
[設定値及び初期値 ]
•name
• [設定値 ] : コミュニティ名  (16 文字以内  )
• [初期値 ] : public
[説明 ]
SNMPv2c によるアクセスモードが読み出し専用であるコミュニティ名を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.8 SNMPv2c の読み書き可能なコミュニティ名の設定
[書式 ]
snmpv2c  community  read-write  name
no snmpv2c  community  read-write
[設定値及び初期値 ]
•name
• [設定値 ] : コミュニティ名  (16 文字以内  )
• [初期値 ] : -
[説明 ]
SNMPv2c によるアクセスモードが読み書き可能であるコミュニティ名を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.9 SNMPv2c トラップの送信先の設定
[書式 ]
snmpv2c  trap host host [type [community ]]
no snmpv2c  trap host host
[設定値及び初期値 ]
•host
• [設定値 ] : SNMPv2c トラップの送信先ホストの  IP アドレス  (IPv4/IPv6)
• [初期値 ] : -
•type : メッセージタイプ
• [設定値 ] :
設定値 説明
trap トラップを送信するコマンドリファレンス  | SNMP の設定  | 381

設定値 説明
inform Informリクエストを送信する
• [初期値 ] : trap
•community
• [設定値 ] : コミュニティ名  (16 文字以内  )
• [初期値 ] : -
[説明 ]
SNMPv2c トラップを送信するホストを指定する。コマンドを複数設定することで、複数のホストを同時に指定でき
る。トラップ送信時のコミュニティ名にはこのコマンドの  community  パラメータが用いられるが、省略されている
場合には  snmpv2c trap community  コマンドの設定値が用いられる。
type パラメータで  'inform' を指定した場合は、送信先からの応答があるまで、 5 秒間隔で最大  3 回再送する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.10 SNMPv2c トラップのコミュニティ名の設定
[書式 ]
snmpv2c  trap community  name
no snmpv2c  trap community
[設定値及び初期値 ]
•name
• [設定値 ] : コミュニティ名  (16 文字以内  )
• [初期値 ] : public
[説明 ]
SNMPv2c トラップを送信する際のコミュニティ名を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.11 SNMPv3 エンジン  ID の設定
[書式 ]
snmpv3  engine  id engine_id
no snmpv3  engine  id
[設定値及び初期値 ]
•engine_id
• [設定値 ] : SNMP エンジン  ID ( 27 文字以内  )
• [初期値 ] : LAN1 の MAC アドレス
[説明 ]
SNMP エンジンを識別するためのユニークな  ID を設定する。 SNMP エンジン  ID は SNMPv3 通信で相手先に通知さ
れる。
相手先に通知されるフォーマットは以下。
•engine_id  が初期値の場合
「8000049e03 」＋  ( LAN1 の MAC アドレス  )
•engine_id  に任意の値を設定した場合
「8000049e04 」＋  設定値の  ASCII 文字列
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.12 SNMPv3 コンテキスト名の設定
[書式 ]
snmpv3  context  name  name382 | コマンドリファレンス  | SNMP の設定

no snmpv3  context  name
[設定値及び初期値 ]
•name
• [設定値 ] : SNMP コンテキスト名  (16 文字以内  )
• [初期値 ] : -
[説明 ]
SNMP コンテキストを識別するための名前を設定する。 SNMP コンテキスト名は  SNMPv3 通信で相手先に通知され
る。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.13 SNMPv3 USM で管理するユーザの設定
[書式 ]
snmpv3  usm user user_id  name  [group  group_id ] [auth auth_pass  [priv priv_pass ]]
no snmpv3  usm user user_id
[設定値及び初期値 ]
•user_id
• [設定値 ] : ユーザ番号  (1..65535)
• [初期値 ] : -
•name
• [設定値 ] : ユーザ名  (32 文字以内  )
• [初期値 ] : -
•group_id
• [設定値 ] : ユーザグループ番号  (1..65535)
• [初期値 ] : -
•auth : 認証アルゴリズム
• [設定値 ] :
設定値 説明
md5 HMAC-MD5-96
sha HMAC-SHA1-96
• [初期値 ] : -
•auth_pass
• [設定値 ] : 認証パスワード  (8 文字以上、 32 文字以内  )
• [初期値 ] : -
•priv : 暗号アルゴリズム
• [設定値 ] :
設定値 説明
des-cbc DES-CBC
aes128-cfb AES128-CFB
• [初期値 ] : -
•priv_pass
• [設定値 ] : 暗号パスワード  (8 文字以上、 32 文字以内  )
• [初期値 ] : -
[説明 ]
SNMPv3 によるアクセスが可能なユーザ情報を設定する。
ユーザグループ番号を指定した場合は  V ACM によるアクセス制御の対象となる。指定しない場合、そのユーザはす
べての  MIB オブジェクトにアクセスできる。
SNMPv3 では通信内容の認証および暗号化が可能であり、本コマンドでユーザ名と共にアルゴリズムおよびパスワ
ードを設定して使用する。なお、認証を行わず暗号化のみを行うことはできない。コマンドリファレンス  | SNMP の設定  | 383

認証や暗号化の有無、アルゴリズムおよびパスワードは、対向となる  SNMP マネージャ側のユーザ設定と一致させ
ておく必要がある。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.14 SNMPv3 によるアクセスを許可するホストの設定
[書式 ]
snmpv3  host host user user_id  ...
snmpv3  host none
no snmpv3  host [host]
[設定値及び初期値 ]
•host : SNMPv3 によるアクセスを許可するホスト
• [設定値 ] :
設定値 説明
ip_address 1個の IPアドレスまたは間にハイフン (-)をはさんだ
IPアドレス (範囲指定 )
lanN LANインターフェース名
bridgeN ブリッジインターフェース名
any すべてのホストからのアクセスを許可する
• [初期値 ] : -
• none : すべてのホストからのアクセスを禁止する
• [初期値 ] : none
•user_id  : ユーザ番号
• [設定値 ] :
• 1 個の数字、または間に  - をはさんだ数字  ( 範囲指定  )、およびこれらを任意に並べたもの  (128 個以内  )
• [初期値 ] : -
[説明 ]
SNMPv3 によるアクセスを許可するホストを設定する。
host パラメータに  'any' を設定した場合は任意のホストからの  SNMPv3 によるアクセスを許可する。なお、アクセス
のあったホストが  host パラメータに合致していても、 user_id  パラメータで指定したユーザに合致しなければアクセ
スはできない。
[ノート ]
host パラメーターにおける  IP アドレスの範囲指定と  LAN インターフェース名の指定は、 RTX5000 / RTX3500
Rev.14.00.17 以前では不可能。ブリッジインターフェースの指定は、 vRX Amazon EC2 版、 および  RTX5000 / RTX3500
Rev.14.00.17 以前では不可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.15 SNMPv3 V ACM で管理する  MIB ビューファミリの設定
[書式 ]
snmpv3  vacm  view  view_id  type oid [type oid ...]
no snmpv3  vacm  view  view_id
[設定値及び初期値 ]
•view_id
• [設定値 ] : ビュー番号  (1..65535)
• [初期値 ] : -
•type
• [設定値 ] :384 | コマンドリファレンス  | SNMP の設定

設定値 説明
include 指定したオブジェクト  ID を管理対象にする
exclude 指定したオブジェクト  ID を管理対象から除外する
• [初期値 ] : -
•oid
• [設定値 ] : MIB オブジェクト  ID (サブ  ID の数は  2 個以上、 128 個以下 )
• [初期値 ] : -
[説明 ]
V ACM による管理で使用する  MIB ビューファミリを設定する。 MIB ビューファミリとは、 アクセス権を許可する際
に指定する  MIB 変数の集合である。
type パラメータと  oid パラメータの組は、 指定のオブジェクト  ID 以降の  MIB サブツリーを管理対象とする／しない
ことを意味する。また複数の組を指定した際に、それぞれ指定したオブジェクト  ID の中で包含関係にあるものは、
より下位の階層まで指定したオブジェクト  ID に対応する  type パラメータが優先される。 128 組まで指定可能。
[設定例 ]
• inetnet サブツリー  (1.3.6.1) 以降を管理対象とする。ただし  enterprises サブツリー  (1.3.6.1.4.1) 以降は管理対象か
ら除外する
# snmpv3 vacm view 1 include 1.3.6.1 exclude 1.3.6.1.4.1
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.16 SNMPv3 V ACM で管理するアクセスポリシーの設定
[書式 ]
snmpv3  vacm  access  group_id  read  read_view  write  write_view
no snmpv3  vacm  access  group_id
[設定値及び初期値 ]
•group_id
• [設定値 ] : グループ番号  (1..65535)
• [初期値 ] : -
•read_view
• [設定値 ] :
設定値 説明
view_id 読み出し可能なアクセス権を設定するビュー番号
none 読み出し可能なビューを設定しない
• [初期値 ] : -
•write_view
• [設定値 ] :
設定値 説明
view_id 書き込み可能なアクセス権を設定するビュー番号
none 書き込み可能なビューを設定しない
• [初期値 ] : -
[説明 ]
ユーザグループに対してアクセスできる  MIB ビューファミリを設定する。このコマンドで設定された  MIB ビュー
ファミリに含まれない  MIB 変数へのアクセスは禁止される。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830コマンドリファレンス  | SNMP の設定  | 385

21.17 SNMPv3 トラップの送信先の設定
[書式 ]
snmpv3  trap host host [type] user user_id
no snmpv3  trap host host
[設定値及び初期値 ]
•host
• [設定値 ] : SNMPv3 トラップの送信先ホストの  IP アドレス  (IPv4/IPv6)
• [初期値 ] : -
•type : メッセージタイプ
• [設定値 ] :
設定値 説明
trap トラップを送信する
inform Informリクエストを送信する
• [初期値 ] : trap
•user_id
• [設定値 ] : ユーザ番号
• [初期値 ] : -
[説明 ]
SNMPv3 トラップを送信するホストを指定する。コマンドを複数設定することで、複数のホストを同時に指定でき
る。トラップ送信時のユーザ設定は  snmpv3 usm user  コマンドで設定したユーザ設定が用いられる。
type パラメータで  'inform' を指定した場合は、送信先からの応答があるまで、 5 秒間隔で最大  3 回再送する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.18 SNMP 送信パケットの始点アドレスの設定
[書式 ]
snmp  local  address  ip_address
no snmp  local  address
[設定値及び初期値 ]
•ip_address
• [設定値 ] : IP アドレス  (IPv4/IPv6)
• [初期値 ] : インタフェースに設定されているアドレスから自動選択
[説明 ]
SNMP 送信パケットの始点  IP アドレスを設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.19 sysContact の設定
[書式 ]
snmp  syscontact  name
no snmp  syscontact
[設定値及び初期値 ]
•name
• [設定値 ] : sysContact として登録する名称  (255 文字以内  )
• [初期値 ] : -
[説明 ]
MIB 変数  sysContact を設定する。空白を含ませるためには、パラメータ全体をダブルクォート  (")、もしくはシング386 | コマンドリファレンス  | SNMP の設定

ルクォート  (') で囲む。
sysContact は一般的に、管理者の名前や連絡先を記入しておく変数である。
[設定例 ]
# snmp syscontact "RT administrator"
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.20 sysLocation の設定
[書式 ]
snmp  syslocation  name
no snmp  syslocation
[設定値及び初期値 ]
•name
• [設定値 ] : sysLocation として登録する名称  (255 文字以内  )
• [初期値 ] : -
[説明 ]
MIB 変数  sysLocation を設定する。空白を含ませるためには、パラメータ全体をダブルクォート  (")、もしくはシング
ルクォート  (') で囲む。
sysLocation は一般的に、機器の設置場所を記入しておく変数である。
[設定例 ]
# snmp syslocation "RT room"
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.21 sysName の設定
[書式 ]
snmp  sysname  name
no snmp  sysname
[設定値及び初期値 ]
•name
• [設定値 ] : sysName として登録する名称  (255 文字以内  )
• [初期値 ] : -
[説明 ]
MIB 変数  sysName を設定する。空白を含ませるためには、パラメータ全体をダブルクォート  (")、もしくはシングル
クォート  (') で囲む。
sysName は一般的に、機器の名称を記入しておく変数である。
[設定例 ]
# snmp sysname "RTX5000 with BRI module"
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.22 SNMP 標準トラップを送信するか否かの設定
[書式 ]
snmp  trap enable  snmp  trap [trap...]
snmp  trap enable  snmp  allコマンドリファレンス  | SNMP の設定  | 387

no snmp  trap enable  snmp
[設定値及び初期値 ]
•trap : 標準トラップの種類
• [設定値 ] :
設定値 説明
coldstart 電源投入時
warmstart 再起動時
linkdown リンクダウン時
linkup リンクアップ時
authenticationfailure 認証失敗時
• [初期値 ] : -
• all : 全ての標準トラップを送信する
• [初期値 ] : -
[初期設定 ]
snmp trap enable snmp all
[説明 ]
SNMP 標準トラップを送信するか否かを設定する。
all を設定した場合には、すべての標準トラップを送信する。個別にトラップを設定した場合には、設定されたトラ
ップだけが送信される。
[ノート ]
authenticationFailure トラップを送信するか否かはこのコマンドによって制御される。
coldStart トラップは、電源投入、再投入による起動後およびファームウェアリビジョンアップによる再起動後に
coldStart トラップを送信する。
linkDown トラップは、 snmp trap send linkdown  コマンドによってインタフェース毎に制御できる。あるインタフェ
ースについて、 linkDown トラップが送信されるか否かは、 snmp trap send linkdown  コマンドで送信が許可されてお
り、かつ、このコマンドでも許可されている場合に限られる。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.23 CPU 使用率監視機能による SNMP トラップを送信するか否かの設定
[書式 ]
snmp  trap cpu threshold  switch
no snmp  trap cpu threshold
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on 送信する
off 送信しない
• [初期値 ] : off
[説明 ]
system cpu threshold により設定した警告を発する CPU 使用率の閾値の上限を超える、または、閾値の下限を下回っ
た際に SNMP トラップを送信するか否かの設定388 | コマンドリファレンス  | SNMP の設定

[ノート ]
RTX5000 、RTX3500 は Rev.14.00.29 以降で使用可能。  RTX1210 は Rev.14.01.34 以降で使用可能。  RTX830 は
Rev.15.02.10 以降で使用可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.24 メモリ使用率監視機能による SNMP トラップを送信するか否かの設定
[書式 ]
snmp  trap memory  threshold  switch
no snmp  trap memory  threshold
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on 送信する
off 送信しない
• [初期値 ] : off
[説明 ]
system memory threshold により設定した警告を発するメモリ使用率の閾値の上限を超える、または、閾値の下限を下
回った際に SNMP トラップを送信するか否かの設定
[ノート ]
RTX5000 、RTX3500 は Rev.14.00.29 以降で使用可能。  RTX1210 は Rev.14.01.34 以降で使用可能。  RTX830 は
Rev.15.02.10 以降で使用可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.25 SNMP トラップの送信の遅延時間の設定
[書式 ]
snmp  trap delay-timer  [wait]
snmp  trap delay-timer  off
no snmp  trap delay-timer  [wait]
[設定値及び初期値 ]
•wait
• [設定値 ] : SNMP トラップを送信するまでの遅延時間の秒数 (1 .. 21474836)
• [初期値 ] : -
[説明 ]
SNMPトラップを送信するイベントが発生してからトラップを送信するまでの間隔を指定する。 offを設定した場
合、即座に SNMPトラップを送信する。設定する遅延時間は最低限保証する値であり、設定値以上遅延する場合も
ある。
[ノート ]
RTX5000 は Rev.14.00.26 以降で使用可能。
RTX3500 は Rev.14.00.26 以降で使用可能。
RTX1210 はRev.14.01.28 以降で使用可能。
RTX830はRev.15.02.03 以降で使用可能。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830コマンドリファレンス  | SNMP の設定  | 389

21.26 SNMP の linkDown トラップの送信制御の設定
[書式 ]
snmp  trap send  linkdown  interface  switch
snmp  trap send  linkdown  pp peer_num  switch
snmp  trap send  linkdown  tunnel  tunnel_num  switch
no snmp  trap send  linkdown  interface
no snmp  trap send  linkdown  pp peer_num
no snmp  trap send  linkdown  tunnel  tunnel_num
[設定値及び初期値 ]
•interface
• [設定値 ] :
• LAN インタフェース名
• WAN インタフェース名
• vRX シリーズ、 RTX5000 、および  RTX3500 では指定不可能
• BRI インタフェース名
• RTX5000 、RTX3500 、および  RTX1210 で指定可能
• [初期値 ] : -
•peer_num
• [設定値 ] : 相手先情報番号
• [初期値 ] : -
•tunnel_num
• [設定値 ] : トンネルインタフェース番号
• [初期値 ] : -
•switch
• [設定値 ] :
設定値 説明
on 送信する
off 送信しない
• [初期値 ] : on
[説明 ]
指定したインタフェースの  linkDown トラップを送信するか否かを設定する。
[拡張ライセンス対応 ]
拡張ライセンス をインポートすると、以下のパラメーターに入力できる上限値が拡張される。
•tunnel_num
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 300 500 700 900 1100
YSL-VPN-EX3 1500 2000 2500 3000 -
•peer_num
ライセンス名拡張後の上限値
( ライセンス本数ごとの値  )
1本 2本 3本 4本 5本
YSL-VPN-EX1 100 - - - -
YSL-VPN-EX2 150390 | コマンドリファレンス  | SNMP の設定

[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.27 PP インタフェースの情報を  MIB2 の範囲で表示するか否かの設定
[書式 ]
snmp  yrifppdisplayatmib2  switch
no snmp  yrifppdisplayatmib2
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on MIB 変数  yrIfPpDisplayAtMib2 を "enabled(1)" とする
off MIB 変数  yrIfPpDisplayAtMib2 を "disabled(2)" とする
• [初期値 ] : off
[説明 ]
MIB 変数  yrIfPpDisplayAtMib2 の値をセットする。この  MIB 変数は、 PP インタフェースを  MIB2 の範囲で表示する
かどうかを決定する。 Rev.4 以前と同じ表示にする場合には、 MIB 変数を  "enabled(1)" に、つまり、このコマンドで
on を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.28 トンネルインタフェースの情報を  MIB2 の範囲で表示するか否かの設定
[書式 ]
snmp  yriftunneldisplayatmib2  switch
no snmp  yriftunneldisplayatmib2
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on MIB 変数  yrIfTunnelDisplayAtMib2 を "enabled(1)" と
する
off MIB 変数  yrIfTunnelDisplayAtMib2 を "disabled(2)" と
する
• [初期値 ] : off
[説明 ]
MIB 変数  yrIfTunnelDisplayAtMib2 の値をセットする。この  MIB 変数は、トンネルインタフェースを  MIB2 の範囲で
表示するかどうかを決定する。 Rev.4 以前と同じ表示にする場合には、 MIB 変数を  "enabled(1)" に、つまり、このコ
マンドで  on を設定する。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.29 スイッチのインタフェースの情報を  MIB2 の範囲で表示するか否かの設定
[書式 ]
snmp  yrifswitchdisplayatmib2  switch
no snmp  yrifswitchdisplayatmib2
[設定値及び初期値 ]
•switch
• [設定値 ] :コマンドリファレンス  | SNMP の設定  | 391

設定値 説明
on MIB 変数  yrIfSwitchDisplayAtMib2 を "enabled(1)" と
する
off MIB 変数  yrIfSwitchDisplayAtMib2 を "disabled(2)" と
する
• [初期値 ] : on
[説明 ]
MIB 変数  yrIfSwitchDisplayAtMib2 の値をセットする。この  MIB 変数は、スイッチのインタフェースを  MIB2 の範囲
で表示するかどうかを決定する。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.30 PP インタフェースのアドレスの強制表示の設定
[書式 ]
snmp  display  ipcp force  switch
no snmp  display  ipcp force
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on IPCP により付与された  IP アドレスを  PP インタフェ
ースのアドレスとして必ず表示する
off IPCP により付与された  IP アドレスは  PP インタフェ
ースのアドレスとして必ずしも表示されない
• [初期値 ] : off
[説明 ]
NAT を使用しない場合や、 NAT の外側アドレスとして固定の  IP アドレスが指定されている場合には、 IPCP で得ら
れた  IP アドレスはそのまま  PP インタフェースのアドレスとして使われる。この場合、 SNMP では通常のインタフ
ェースの  IP アドレスを調べる手順で  IPCP としてどのようなアドレスが得られたのか調べることができる。
しかし、 NAT の外側アドレスとして  'ipcp' と指定している場合には、 IPCP で得られた  IP アドレスは  NAT の外側ア
ドレスとして使用され、インタフェースには付与されない。そのため、 SNMP でインタフェースの  IP アドレスを調
べても、 IPCP でどのようなアドレスが得られたのかを知ることができない。
本コマンドを  on に設定しておくと、 IPCP で得られた  IP アドレスが  NAT の外側アドレスとして使用される場合で
も、 SNMP ではそのアドレスをインタフェースのアドレスとして表示する。アドレスが実際にインタフェースに付
与されるわけではないので、始点  IP アドレスとして、その  IP アドレスが利用されることはない。
[適用モデル ]
vRX シリーズ , RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.31 LAN インタフェースの各ポートのリンクが  up/down したときにトラップを送信するか
否かの設定
[書式 ]
snmp  trap link-updown  separate-l2switch-port  interface  switch
no snmp  trap link-updown  separate-l2switch-port  interface
[設定値及び初期値 ]
•interface
• [設定値 ] : スイッチングハブを持つ  LAN インタフェース名
• [初期値 ] : -
•switch
• [設定値 ] :392 | コマンドリファレンス  | SNMP の設定

設定値 説明
on トラップを送信する
off トラップを送信しない
• [初期値 ] : off
[説明 ]
各ポートのリンクが  up/down したときにトラップを送信するか否かを設定する。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.32 電波強度トラップを送信するか否かの設定
[書式 ]
snmp  trap mobile  signal-strength  switch  [level ]
no snmp  trap mobile  signal-strength  [switch  [level ]]
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on トラップを送信する
off トラップを送信しない
• [初期値 ] : off
•level  : アンテナ本数の閾値
• [設定値 ] :
設定値 説明
0..3 アンテナ本数
省略 省略時は圏外
• [初期値 ] : -
[説明 ]
モバイル端末の電波強度トラップを送信するか否かを設定する。  自動 /手動に関わらず、ルータが電波強度を取得し
た時にトラップ送信が許可されており、  電波強度のアンテナ本数が閾値以下であった場合にトラップが送信される。
[ノート ]
トラップは yrIfMobileStatusTrap が送信される。
[適用モデル ]
RTX3510, RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.33 スイッチへ静的に付与するインタフェース番号の設定
[書式 ]
snmp  ifindex  switch  static  index  index  switch
no snmp  ifindex  switch  static  index  index  [switch ]
[設定値及び初期値 ]
•index
• [設定値 ] : オブジェクト IDのインデックス (100000000 .. 199999999)
• [初期値 ] : -
•switch  : MAC アドレス、あるいはポート番号の組
• [初期値 ] : -
[説明 ]
スイッチのインタフェースを示すオブジェクト IDのインデックスの先頭を静的に指定する。コマンドリファレンス  | SNMP の設定  | 393

[ノート ]
オブジェクト  ID が重複した場合の動作は保証されない。
静的にオブジェクト  ID のインデックスの先頭を指定した場合、スイッチのインタフェースを示すオブジェクト  ID
のインデックスは動的に割り当てられない。
snmp yrswindex switch static index  コマンドが設定された場合、 snmp yrswindex switch static index  コマンドで指定さ
れたスイッチのみインデックスが割り当てられる。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.34 スイッチへ静的に付与するスイッチ番号の設定
[書式 ]
snmp  yrswindex  switch  static  index  index  switch
no snmp  yrswindex  switch  static  index  index  [switch ]
[設定値及び初期値 ]
•index
• [設定値 ] : オブジェクト IDのインデックス (1 .. 2147483647)
• [初期値 ] : -
•switch  : MAC アドレス、あるいはポート番号の組
• [初期値 ] : -
[説明 ]
スイッチのオブジェクト IDのインデックスを静的に指定する。
[ノート ]
静的にオブジェクト IDのインデックスを指定した場合、スイッチのオブジェクト IDのインデックスは動的に割り
当てられない。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.35 スイッチの状態による  SNMP トラップの条件の設定
[書式 ]
snmp  trap enable  switch  switch  trap [trap...]
snmp  trap enable  switch  switch  all
snmp  trap enable  switch  switch  none
no snmp  trap enable  switch  switch
[設定値及び初期値 ]
•switch  : default、MAC アドレス、あるいはポート番号の組
• [初期値 ] : -
•trap : トラップの種類
• [設定値 ] :
設定値 説明
linkup リンクアップ時
linkdown リンクダウン時
fanlock ファン異常時
loopdetect ループ検出時
poesuppply 給電開始
poeterminate 給電停止
oversupply 給電能力オーバー394 | コマンドリファレンス  | SNMP の設定

設定値 説明
overtemp 温度異常
powerfailure 電源異常
• [初期値 ] : -
• all : 全てのトラップを送信する
• [初期値 ] : -
• none : 全てのトラップを送信しない
• [初期値 ] : -
[初期設定 ]
snmp trap enable switch default all
[説明 ]
選択されたスイッチの監視状態に応じてトラップを送信する条件を設定する。 defaultを指定して設定した場合は、
個別のスイッチについて  SNMP トラップの条件の設定がない場合の動作を決定する。
all を設定した場合には、すべてのトラップを送信する。 none を設定した場合には、すべてのトラップを送信しな
い。個別にトラップを設定した場合には、設定されたトラップだけが送信される。
リンクアップ・リンクダウントラップは標準 MIBのトラップであり、送信するには  snmp trap enable snmp  コマン
ドでもトラップ送信が許可されている必要がある。
ループ検出のトラップを送信するにはスイッチ側に  switch control function set loopdetect-linkdown linkdown  コマン
ドあるいは  switch control function set loopdetect-linkdown linkdown-recovery  コマンドが設定されている必要があ
る。
給電開始、給電停止、給電能力オーバー、温度異常、電源異常のトラップを設定した場合、 SWX2200-8PoE 以外の
スイッチではトラップは送信されない。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830
21.36 スイッチで共通の  SNMP トラップの条件の設定
[書式 ]
snmp  trap enable  switch  common  trap [trap...]
snmp  trap enable  switch  common  all
snmp  trap enable  switch  common  none
no snmp  trap enable  switch  common
[設定値及び初期値 ]
•trap : トラップの種類
• [設定値 ] :
設定値 説明
find-switch スイッチが監視下に入った時
detect-down スイッチが監視から外れた時
• [初期値 ] : -
• all : 全てのトラップを送信する
• [初期値 ] : -
• none : 全てのトラップを送信しない
• [初期値 ] : -
[初期設定 ]
snmp trap enable switch common all
[説明 ]
スイッチの監視状態に応じてトラップを送信する条件を設定する。コマンドリファレンス  | SNMP の設定  | 395

[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830396 | コマンドリファレンス  | SNMP の設定

