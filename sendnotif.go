// Package sendnotif is simple implementation for sending desktop notifications
// It uses godbus library
// Specification: https://specifications.freedesktop.org/notification/1.3/index.html
package sendnotif

import (
	"errors"
	"math"
	"time"

	"github.com/godbus/dbus/v5"
)

const (
	notifObjDest = "org.freedesktop.Notifications"
	notifObjPath = "/org/freedesktop/Notifications"

	notifNotifyMethod = notifObjDest + ".Notify"
	notifCloseMethod  = notifObjDest + ".CloseNotification"
)

const (
	ExpireInfinite time.Duration = -1            // infinite notification expiration, Note: depends on notification server, may not work
	ExpireMax      time.Duration = math.MaxInt64 // maximum duration for notification expiration
)

var (
	ErrConn        = errors.New("unable to connect to session bus")
	ErrNotifiSend  = errors.New("unable to send notification")
	ErrNotifiClose = errors.New("unable to close notification")
)

// Notifications holds the connection to dbus session bus
type Notifications struct {
	conn *dbus.Conn
}

// Init() connects to notification server over session dbus, call it before any operations
func Init() (Notifications, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return Notifications{}, ErrConn
	}

	return Notifications{
		conn: conn,
	}, nil
}

// TODO: GetCapabilities for actions and more

// Close() closes connection to dbus bus, call it after you are done
func (n *Notifications) Close() {
	n.conn.Close()
}

// Notification holds all data of notification
type Notification struct {
	appName    string // app_name
	replacesId uint32 // replaces_id
	appIcon    string // app_icon
	summary    string // summary
	body       string // body
	// actions    []string // actions, check if implemented
	// map[string]dbus.Variant{}, // hints
	expireTimeout int32 // expire_timeout
}

// NewNotification() creates and returns new notification, nothing is send to server
func NewNotification(appName, summary, body string, expire time.Duration) Notification {
	return Notification{
		appName: appName,
		// replacesId: 0,
		// appIcon: "",
		summary:       summary,
		body:          body,
		expireTimeout: int32(expire.Milliseconds()),
	}
}

// Send() sends prepared notification and returns back Notification with replacesId
func (n Notifications) Send(notification Notification) (Notification, error) {
	var res dbus.Variant

	obj := n.conn.Object(notifObjDest, notifObjPath)
	call := obj.Call(
		notifNotifyMethod, 0,
		notification.appName,
		notification.replacesId,
		notification.appIcon,
		notification.summary,
		notification.body,
		[]string{},                // TODO: actions
		map[string]dbus.Variant{}, // TODO: hints
		notification.expireTimeout,
	)
	call.Store(&res)

	if call.Err != nil {
		return notification, ErrNotifiSend
	} else {
		var id uint32
		if err := res.Store(&id); err != nil {
			return notification, ErrNotifiSend
		}

		notification.replacesId = id
		return notification, nil
	}
}

// GetReplacesId() returns replacesId of Notification
func (n Notification) GetReplacesId() uint32 {
	return n.replacesId
}

// SetReplacesId() sets replacesId on Notification
func (n *Notification) SetReplacesId(replacesId uint32) {
	n.replacesId = replacesId
}

// CloseNotification() close Notification (based of replacesId)
func (n Notifications) CloseNotification(notification Notification) error {
	obj := n.conn.Object(notifObjDest, notifObjPath)
	call := obj.Call(
		notifCloseMethod, 0,
		notification.replacesId,
	)

	if call.Err != nil {
		return ErrNotifiClose
	}

	return nil
}

/*
<node>
  <interface name="org.freedesktop.notifications">
    <method name="setnotiwindowvisibility">
      <arg type="b" name="value" direction="in"/>
    </method>

    <method name="manuallyclosenotification">
      <arg type="u" name="id" direction="in"/>
      <arg type="b" name="timeout" direction="in"/>
    </method>

    <method name="hidelatestnotification">
      <arg type="b" name="close" direction="in"/>
    </method>

    <method name="notify">
      <arg type="s" name="app_name" direction="in"/>
      <arg type="u" name="replaces_id" direction="in"/>
      <arg type="s" name="app_icon" direction="in"/>
      <arg type="s" name="summary" direction="in"/>
      <arg type="s" name="body" direction="in"/>
      <arg type="as" name="actions" direction="in"/>
      <arg type="a{sv}" name="hints" direction="in"/>
      <arg type="i" name="expire_timeout" direction="in"/>
      <arg type="u" name="result" direction="out"/>
    </method>

    <method name="closenotification">
      <arg type="u" name="id" direction="in"/>
    </method>

    <signal name="notificationclosed">
      <arg type="u" name="id"/>
      <arg type="u" name="reason"/>
    </signal>

    <signal name="actioninvoked">
      <arg type="u" name="id"/>
      <arg type="s" name="action_key"/>
    </signal>

    <signal name="activationtoken">
      <arg type="u" name="id"/>
      <arg type="s" name="activation_token"/>
    </signal>

    <signal name="notificationreplied">
      <arg type="u" name="id"/>
      <arg type="s" name="text"/>
    </signal>
  </interface>
</node>
*/
