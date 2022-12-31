
#################################################################################################
# Check Result file
#################################################################################################
cat ./result.json

echo ""
echo ""
echo "Checking result file ..."
echo ""
if [[ $(jq '. | length' result.json) != 1 ]]; then
  echo "Number of results not equals 1"; exit 1;
fi

if [[ $(jq '.["nginx:1.14.2"].Findings | length' result.json) != 3 ]]; then
  echo "Number of Findings for \"nginx:1.14.2\" not equals 3"; exit 1;
fi

if [[ "$(jq -r '.["nginx:1.14.2"].Findings[0].NotificationData.Email' result.json)" != "ns-admin@org.com" ]]; then
  echo "First finding for \"nginx:1.14.2\": Email value not expected"; exit 1;
fi
echo "Finished! All ok ..."


#################################################################################################
# Check Emails
#################################################################################################
echo ""
echo ""
echo "Checking result emails ..."

### Admin Email
echo ""
echo "Checking Admin email ..."
echo ""
adminEmailIndex=1
curl -s localhost:8025/api/v1/messages | jq ".[0]"
echo ""
if [[ "$(curl -s localhost:8025/api/v1/messages | jq -r ".[$adminEmailIndex].To[0].Mailbox")" != "nniikkoollaaii" ]]; then
  echo "Check Admin email recipient failed"; exit 1;
fi
if [[ "$(curl -s localhost:8025/api/v1/messages | jq -r ".[$adminEmailIndex].From.Mailbox")" != "test" ]]; then
  echo "Check Admin email sender failed"; exit 1;
fi

if [[ "$(curl -s localhost:8025/api/v1/messages | jq -r ".[$adminEmailIndex].Content.Headers.Subject[0]")" != "Outdated container images older than 0h in use [Admin Report]" ]]; then
  echo "Check Admin email subject failed"; exit 1;
fi


curl -s localhost:8025/api/v1/messages | jq -r ".[$adminEmailIndex].Content.Body" | tr -d '\r\n' > test/email-admin-body-content.actual.txt
cat -n test/email-admin-body-content.actual.txt
cat -n test/email-admin-body-content.expected.txt

if [[ "$(cat -n test/email-admin-body-content.actual.txt)" != "$(cat -n test/email-admin-body-content.expected.txt)" ]]; then
  echo "Check Admin email body content failed"; exit 1;
fi
echo "Admin Email Finished! All ok ..."


### User email
echo ""
echo "Checking User email ..."
echo ""
userEmailIndex=0
curl -s localhost:8025/api/v1/messages | jq ".[0]"
echo ""
if [[ "$(curl -s localhost:8025/api/v1/messages | jq -r ".[$userEmailIndex].To[0].Mailbox")" != "ns-admin" ]]; then
  echo "Check User email recipient failed"; exit 1;
fi
if [[ "$(curl -s localhost:8025/api/v1/messages | jq -r ".[$userEmailIndex].From.Mailbox")" != "test" ]]; then
  echo "Check User email sender failed"; exit 1;
fi

if [[ "$(curl -s localhost:8025/api/v1/messages | jq -r ".[$userEmailIndex].Content.Headers.Subject[0]")" != "Outdated container images older than 0h in use" ]]; then
  echo "Check User email subject failed"; exit 1;
fi


curl -s localhost:8025/api/v1/messages | jq -r ".[$userEmailIndex].Content.Body" | tr -d '\r\n' > test/email-user-body-content.actual.txt
if [[ "$(cat -n test/email-user-body-content.actual.txt)" != "$(cat -n test/email-user-body-content.expected.txt)" ]]; then
  echo "Check User email body content failed"; exit 1;
fi
echo "User Email Finished! All ok ..."
