use strict;
use warnings;
use Test::More;
use LWP::UserAgent;
use HTTP::Request::Common qw(POST GET PUT);
use HTTP::Cookies;
use File::Slurp;
use Data::Dumper;
use JSON::PP qw(encode_json decode_json);

my $BASE_URL = "http://localhost:8080";
my $ua = LWP::UserAgent->new(cookie_jar => HTTP::Cookies->new());

sub generate_username {
    my $random_string = join '', map { ('a'..'z', 'A'..'Z', 0..9)[rand 62] } 1..8;
    my $timestamp = sprintf("%.6f", time);
    return "user_${random_string}_$timestamp";
}

sub extract_session_cookie {
    my ($response) = @_;
    my $cookies = $response->header('Set-Cookie');
    my ($session_cookie) = $cookies =~ /(session-name=[^;]+)/;
    return $session_cookie;
}

sub debug_json {
    my $response = shift;
    print JSON::PP->new->pretty->encode(decode_json($response->content));
}

sub test_user_flow {
    my $username = generate_username();

    # Create new user
    my $create_user_response = $ua->request(
        POST "$BASE_URL/api/users",
        Content_Type => 'application/json',
        Content => encode_json({
            username => $username,
            password => "123456",
            role => "user"
        })
    );
    is($create_user_response->code, 201, "Create user successful");

    # Attempt to create the same user again (should fail)
    my $duplicate_user_response = $ua->request(
        POST "$BASE_URL/api/users",
        Content_Type => 'application/json',
        Content => encode_json({
            username => $username,
            password => "123456",
            role => "user"
        })
    );
    
    is($duplicate_user_response->code, 409, "Duplicate user creation failed as expected");
    my $error_message = decode_json($duplicate_user_response->content)->{message};
    is($error_message, "Username already exists", "Correct error message for duplicate username");

    # Login
    my $login_response = $ua->request(
        POST "$BASE_URL/login",
        Content_Type => 'application/json',
        Content => encode_json({
            username => $username,
            password => "123456"
        })
    );
    is($login_response->code, 200, "Login successful");
    
    # Extract session cookie
    my $session_cookie = extract_session_cookie($login_response);
    ok($session_cookie, "Session cookie extracted");

    # Check unauthorized access
    my $unauthorized_ua = LWP::UserAgent->new();
    my $unauthorized_response = $unauthorized_ua->get("$BASE_URL/login/check");
    is($unauthorized_response->code, 401, "Unauthorized access denied");

    # Check authorized access
    my $authorized_response = $ua->get("$BASE_URL/login/check",
        'Cookie' => $session_cookie
    );
    is($authorized_response->code, 200, "Authorized access successful");

    return $session_cookie;
}

sub test_picture_upload_and_revisions {
    my ($session_cookie) = @_;
    
    # Upload first revision for the new art project
    my $upload_response = $ua->request(
        POST "$BASE_URL/self/artprojects",
        'Cookie' => $session_cookie,
        Content_Type => 'form-data',
        Content => [
            file => ['data/1.png'],
            title => 'This is the one'
        ]
    );

    is($upload_response->code, 201, "Art project with first revision successfully created");
    my $upload_data = decode_json($upload_response->content);
    ok(exists $upload_data->{id}, "Art-project ID received");
    my $artproject_id = $upload_data->{id};

    my $art_project_link1 = "$BASE_URL/self/artprojects/$artproject_id";
    my $art_project_link_resp = $ua->request(
        GET $art_project_link1,
        'Cookie' => $session_cookie,
    );
    is($art_project_link_resp->code, 200, "Get art-project datae: $art_project_link1");
    # debug_json($art_project_link_resp);

    # Add second revision to the first art project
    my $revision_response = $ua->request(
        POST "$BASE_URL/self/artprojects/$artproject_id/revisions",
        'Cookie' => $session_cookie,
        Content_Type => 'form-data',
        Content => [
            file => ['data/2.png'],
            comment => 'second version'
        ]
    );

    is($revision_response->code, 201, "First picture revision successful");
    my $revision_data = decode_json($revision_response->content);
    is($revision_data->{version}, 2, "Revision version is 2");

    my $art_link = "$BASE_URL/art/".$revision_data->{art_id};
    my $revision_public_link = $ua->request(
        GET $art_link
    );
    is($revision_public_link->code, 200, "Public link '$art_link' to the revision 2 is reachable");
    is($revision_public_link->header('Content-Type'), "image/png", "Content type is correct");

    # Add third revision to the first art project
    $revision_response = $ua->request(
        POST "$BASE_URL/self/artprojects/$artproject_id/revisions",
        'Cookie' => $session_cookie,
        Content_Type => 'form-data',
        Content => [
            file => ['data/3.png'],
            comment => 'third version'
        ]
    );
    
    is($revision_response->code, 201, "Third art revision upload successful");
    $revision_data = decode_json($revision_response->content);
    is($revision_data->{version}, 3, "Revision version is 3");

    # Create second art project
    my $upload_response2 = $ua->request(
        POST "$BASE_URL/self/artprojects",
        'Cookie' => $session_cookie,
        Content_Type => 'form-data',
        Content => [
            file => ['data/nun01.jpeg'],
            title => 'AI generated nun'
        ]
    );
    is($upload_response2->code, 201, "Second art project with first revision successfully created");
    my $upload_data2 = decode_json($upload_response2->content);
    my $picture_id2 = $upload_data2->{id};

    # Add first revision to second picture
    my $revision_response2 = $ua->request(
        POST "$BASE_URL/self/artprojects/$picture_id2/revisions",
        'Cookie' => $session_cookie,
        Content_Type => 'form-data',
        Content => [
            file => ['data/nun02.jpeg']
        ]
    );
    is($revision_response2->code, 201, "Second art project second revision successful");
    my $revision_data2 = decode_json($revision_response2->content);
    is($revision_data2->{version}, 2, "Revision version is 2 for second art project");

    my $art_link2 = "$BASE_URL/art/".$revision_data2->{art_id};
    my $revision_public_link2 = $ua->request(
        GET $art_link2
    );
    is($revision_public_link2->code, 200, "Public link '$art_link2' to the revision 2 is reachable");

    my $art_projects = $ua->request(
        GET "$BASE_URL/self/artprojects",
        'Cookie' => $session_cookie,
    );

    is($art_projects->code, 200, "Listed art project data successful");
    my $art_projects_data = decode_json($art_projects->content);
    is(@$art_projects_data, 2, "Listed 2 art projects");
    is($art_projects_data->[0]->{title}, "This is the one", "Title of the first art project is correct");
    is($art_projects_data->[1]->{title}, "AI generated nun", "Title of the second art project is correct");

    # create collection: collection_id 
    my $new_collection = $ua->request(
        POST "http://localhost:8080/self/collections",
        'Cookie' => $session_cookie,
        Content_Type => 'application/json',
        Content => encode_json({ title => "Main" })
    );

    is($new_collection->code, 201, "New collection is created successful");
    my $new_collection_data = decode_json($new_collection->content);
    is($new_collection_data->{title}, "Main", "Collection title is correct");
    my $collection_id = $new_collection_data->{id};
    my $public_collection_id = $new_collection_data->{collection_id};

    foreach my $art_project (@$art_projects_data) {
        my $response = $ua->request(
            POST "$BASE_URL/self/collections/$collection_id/revisions",
            'Cookie' => $session_cookie,
            Content_Type => 'application/json',
            Content => encode_json({ revisionID => $art_project->{latest_revision_id} })
        );

        is($response->code, 200, "New revision is added to the collection");
    }

    my $collection01 = $ua->request(
        GET "$BASE_URL/self/collections/$collection_id/revisions",
        'Cookie' => $session_cookie
    );
    is($collection01->code, 200, "Got new collection data");
  
    my $collection01_data = decode_json($collection01->content);
    is(@$collection01_data, 2, "Listed 2 revisions in this collection");
    is($collection01_data->[0]->{version}, 3, "Version of the first revision in the collection is correct");
    is($collection01_data->[1]->{version}, 2, "Version of the second revision in the collection is correct");
  
      my $public_collection = $ua->request(
        GET "$BASE_URL/collection/$public_collection_id",
    );
    is($public_collection->code, 200, "Got public collection data");

    my $public_collection_data = decode_json($public_collection->content);
    is(@$public_collection_data, 2, "Listed 2 arts in this public collection");
    is($public_collection_data->[0]->{size}, 8275, "Size of first is correct");
    is($public_collection_data->[1]->{size}, 21966, "Size of second image is correct");

    my $art_project1_latest_revision = $ua->request(
        GET "$BASE_URL/self/artprojects/".$art_projects_data->[0]->{id}."/revisions/".$art_projects_data->[0]->{latest_revision_id},
        'Cookie' => $session_cookie,
    );
    is($art_project1_latest_revision->code, 200, "Latest revision file is found");
    is($art_project1_latest_revision->header('Content-Type'), "image/png", "Content type is correct");
}

my $session_cookie = test_user_flow();
test_picture_upload_and_revisions($session_cookie);

done_testing();