package email

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"sort"
	"strings"
)

type Message struct {
	Subject      string
	Body         string
	From         string
	To           []string
	Cc           []string
	Bcc          []string
	ReplyTo      []string
	Headers      map[string]string
	Alternatives []Alternative
	Attachments  []Attachment
}

type Alternative struct {
	ContentType string
	Body        string
}

type Attachment struct {
	Filename    string
	ContentType string
	Body        []byte
}

func (m *Message) AddAlternative(contentType string, body string) {
	m.Alternatives = append(m.Alternatives, Alternative{ContentType: contentType, Body: body})
}

func (m *Message) Attach(filename string, contentType string, body []byte) {
	m.Attachments = append(m.Attachments, Attachment{Filename: filename, ContentType: contentType, Body: append([]byte(nil), body...)})
}

func (m Message) RenderMIME() ([]byte, error) {
	var body bytes.Buffer
	headers := m.renderHeaders()
	for _, key := range sortedHeaderKeys(headers) {
		if _, err := fmt.Fprintf(&body, "%s: %s\r\n", key, headers.Get(key)); err != nil {
			return nil, err
		}
	}
	if len(m.Attachments) > 0 {
		writer := multipart.NewWriter(&body)
		_, _ = fmt.Fprintf(&body, "Content-Type: multipart/mixed; boundary=%q\r\n\r\n", writer.Boundary())
		if err := m.writeBodyPart(writer); err != nil {
			return nil, err
		}
		for _, attachment := range m.Attachments {
			header := textproto.MIMEHeader{}
			contentType := attachment.ContentType
			if contentType == "" {
				contentType = "application/octet-stream"
			}
			header.Set("Content-Type", contentType)
			header.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, attachment.Filename))
			part, err := writer.CreatePart(header)
			if err != nil {
				return nil, err
			}
			if _, err := part.Write(attachment.Body); err != nil {
				return nil, err
			}
		}
		if err := writer.Close(); err != nil {
			return nil, err
		}
		return body.Bytes(), nil
	}
	if len(m.Alternatives) > 0 {
		writer := multipart.NewWriter(&body)
		_, _ = fmt.Fprintf(&body, "Content-Type: multipart/alternative; boundary=%q\r\n\r\n", writer.Boundary())
		if err := writeTextPart(writer, "text/plain; charset=utf-8", m.Body); err != nil {
			return nil, err
		}
		for _, alternative := range m.Alternatives {
			if err := writeTextPart(writer, alternative.ContentType, alternative.Body); err != nil {
				return nil, err
			}
		}
		if err := writer.Close(); err != nil {
			return nil, err
		}
		return body.Bytes(), nil
	}
	_, _ = fmt.Fprintf(&body, "Content-Type: text/plain; charset=utf-8\r\n\r\n%s", m.Body)
	return body.Bytes(), nil
}

func (m Message) Recipients() []string {
	recipients := append([]string(nil), m.To...)
	recipients = append(recipients, m.Cc...)
	recipients = append(recipients, m.Bcc...)
	return recipients
}

func (m Message) Clone() Message {
	return Message{
		Subject:      m.Subject,
		Body:         m.Body,
		From:         m.From,
		To:           append([]string(nil), m.To...),
		Cc:           append([]string(nil), m.Cc...),
		Bcc:          append([]string(nil), m.Bcc...),
		ReplyTo:      append([]string(nil), m.ReplyTo...),
		Headers:      cloneHeaders(m.Headers),
		Alternatives: append([]Alternative(nil), m.Alternatives...),
		Attachments:  cloneAttachments(m.Attachments),
	}
}

func (m Message) renderHeaders() textproto.MIMEHeader {
	headers := textproto.MIMEHeader{}
	headers.Set("Subject", m.Subject)
	headers.Set("From", m.From)
	if len(m.To) > 0 {
		headers.Set("To", strings.Join(m.To, ", "))
	}
	if len(m.Cc) > 0 {
		headers.Set("Cc", strings.Join(m.Cc, ", "))
	}
	if len(m.ReplyTo) > 0 {
		headers.Set("Reply-To", strings.Join(m.ReplyTo, ", "))
	}
	for key, value := range m.Headers {
		headers.Set(key, value)
	}
	return headers
}

func (m Message) writeBodyPart(writer *multipart.Writer) error {
	if len(m.Alternatives) == 0 {
		return writeTextPart(writer, "text/plain; charset=utf-8", m.Body)
	}
	var nested bytes.Buffer
	alternative := multipart.NewWriter(&nested)
	if err := writeTextPart(alternative, "text/plain; charset=utf-8", m.Body); err != nil {
		return err
	}
	for _, item := range m.Alternatives {
		if err := writeTextPart(alternative, item.ContentType, item.Body); err != nil {
			return err
		}
	}
	if err := alternative.Close(); err != nil {
		return err
	}
	header := textproto.MIMEHeader{}
	header.Set("Content-Type", fmt.Sprintf("multipart/alternative; boundary=%q", alternative.Boundary()))
	part, err := writer.CreatePart(header)
	if err != nil {
		return err
	}
	_, err = part.Write(nested.Bytes())
	return err
}

func writeTextPart(writer *multipart.Writer, contentType string, value string) error {
	if contentType == "" {
		contentType = "text/plain; charset=utf-8"
	}
	header := textproto.MIMEHeader{}
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		return err
	}
	_, err = part.Write([]byte(value))
	return err
}

func sortedHeaderKeys(headers textproto.MIMEHeader) []string {
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func cloneHeaders(headers map[string]string) map[string]string {
	if headers == nil {
		return nil
	}
	cloned := make(map[string]string, len(headers))
	for key, value := range headers {
		cloned[key] = value
	}
	return cloned
}

func cloneAttachments(attachments []Attachment) []Attachment {
	cloned := make([]Attachment, len(attachments))
	for i, attachment := range attachments {
		cloned[i] = Attachment{Filename: attachment.Filename, ContentType: attachment.ContentType, Body: append([]byte(nil), attachment.Body...)}
	}
	return cloned
}
